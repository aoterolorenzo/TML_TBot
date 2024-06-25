package usecases

import (
	"TML_TBot/config"
	"TML_TBot/domain/models"
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"math"
	"os"
)

const SPOILERS_CACHE_FILE = "./.cache/messages.ndjson"

type TMLAntiSpoilersController struct {
	msgID          int
	msgs           map[int]tgbotapi.Message
	responseParams models.ResponseParams
	botToken       string
}

func NewTMLAntiSpoilersController(job models.Job) *TMLAntiSpoilersController {
	msgs, err := readMessagesFromNDJSONFile(SPOILERS_CACHE_FILE)
	if err != nil {
		panic(errors.New(fmt.Sprintf("unable to open file: %s", SPOILERS_CACHE_FILE)))
	}

	return &TMLAntiSpoilersController{
		responseParams: job.Response[0],
		msgID:          math.MinInt64,
		msgs:           msgs,
		botToken:       os.Getenv("TOKEN"),
	}
}

func (t *TMLAntiSpoilersController) Run() ([]models.TGMessage, error) {
	bot, err := tgbotapi.NewBotAPI(t.botToken)
	if err != nil {
		config.Log.Panic(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		config.Log.Panic(err)
	}

	for update := range updates {
		var currentMsg = update.Message
		var topicCheckPassed = false
		if update.Message != nil && update.Message.Chat.ID == int64(t.responseParams.ChatID) {
			t.msgs[update.Message.MessageID] = *update.Message
			err := appendMessageToNDJSONFile(SPOILERS_CACHE_FILE, *update.Message)
			if err != nil {
				return nil, err
			}
		}

		for !topicCheckPassed && currentMsg != nil {
			// if it has father (currentMsg.ReplyToMessage) and ReplyToMessage IS topic
			if currentMsg.ReplyToMessage != nil && currentMsg.ReplyToMessage.MessageID == int(t.responseParams.TopicID) {
				// topicCheckPassed, activate spoiler process
				topicCheckPassed = true
			} else if currentMsg.ReplyToMessage != nil {
				// let's grab the father and continue. if we dont have it, break, topicCheckPassed still false
				if foundFather, exists := t.msgs[currentMsg.ReplyToMessage.MessageID]; exists {
					currentMsg = &foundFather
				} else {
					break
				}
			} else {
				// it doesn't respond to anything. break loop and topicCheckPassed still false
				break
			}
		}

		if update.Message != nil && update.Message.Chat.ID == int64(t.responseParams.ChatID) &&
			topicCheckPassed {
			// If we have a msg id, we just remove the message id
			if t.msgID != math.MinInt64 {
				deleteMessage := tgbotapi.DeleteMessageConfig{
					ChatID:    update.Message.Chat.ID,
					MessageID: t.msgID,
				}

				if _, err := bot.DeleteMessage(deleteMessage); err != nil {
					config.Log.Errorf("Failed to delete message: %v", err)
				}
			}

			// We add a message to the chat, and save its ID
			// Send a new message to the chat
			newMessage := tgbotapi.NewMessage(update.Message.Chat.ID, "🚨<b>SPOILER ALERT</b>")
			newMessage.ParseMode = "HTML"
			if t.responseParams.TopicID != 0 {
				newMessage.ReplyToMessageID = int(t.responseParams.TopicID)
			}

			sentMessage, err := bot.Send(newMessage)
			if err != nil {
				config.Log.Errorf("Failed to send message: %v", err)
				continue
			}

			// Save the sent message's ID in msgID variable
			t.msgID = sentMessage.MessageID
		}
	}

	return nil, nil
}

func readMessagesFromNDJSONFile(filename string) (map[int]tgbotapi.Message, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	messages := make(map[int]tgbotapi.Message)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var msg tgbotapi.Message
		err := json.Unmarshal(scanner.Bytes(), &msg)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal message: %w", err)
		}
		messages[msg.MessageID] = msg
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	return messages, nil
}

func writeMessagesToNDJSONFile(filename string, messages map[int]tgbotapi.Message) error {
	// Open the file in write mode
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, msg := range messages {
		data, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}

		_, err = writer.Write(append(data, '\n'))
		if err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}

	return nil
}

func appendMessageToNDJSONFile(filename string, newMessage tgbotapi.Message) error {
	// Open the file in append mode
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Marshal the new message to JSON
	data, err := json.Marshal(newMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Write the JSON data to the file followed by a newline
	_, err = file.Write(append(data, '\n'))
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}
