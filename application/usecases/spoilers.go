package usecases

import (
	"TML_TBot/config"
	"TML_TBot/domain/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"math"
	"os"
)

type TMLAntiSpoilersController struct {
	msgID          int
	responseParams models.ResponseParams
	botToken       string
}

func NewTMLAntiSpoilersController(job models.Job) *TMLAntiSpoilersController {
	return &TMLAntiSpoilersController{
		responseParams: job.Response[0],
		msgID:          math.MinInt64,
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
		if update.Message != nil && update.Message.Chat.ID == int64(t.responseParams.ChatID) &&
			(t.responseParams.TopicID == 0 || update.Message.ReplyToMessage.MessageID == int(t.responseParams.TopicID)) {
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
