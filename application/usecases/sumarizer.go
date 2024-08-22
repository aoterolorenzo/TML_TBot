package usecases

import (
	"TML_TBot/config"
	"TML_TBot/domain/models"
	"context"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sashabaranov/go-openai" // Ensure you've installed this package
	"io/ioutil"
	"math"
	"net/http"
	"os"
)

type TMLSumarizerController struct {
	msgID          int
	msgs           map[int]tgbotapi.Message
	responseParams models.ResponseParams
	botToken       string
}

func NewTMLSumarizerController(job models.Job) *TMLSumarizerController {
	msgs, err := readMessagesFromNDJSONFile(SPOILERS_CACHE_FILE)
	if err != nil {
		panic(errors.New(fmt.Sprintf("unable to open file: %s", SPOILERS_CACHE_FILE)))
	}

	return &TMLSumarizerController{
		responseParams: job.Response[0],
		msgID:          math.MinInt64,
		msgs:           msgs,
		botToken:       os.Getenv("TOKEN"),
	}
}

func (t *TMLSumarizerController) Run() ([]models.TGMessage, error) {

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
		var udp *tgbotapi.Message
		if update.Message != nil && update.Message.Text == "..." {
			udp = update.Message.ReplyToMessage
		} else {
			continue
		}

		if udp != nil && udp.Chat.ID == int64(t.responseParams.ChatID) {
			var file tgbotapi.File
			if udp.Voice != nil || udp.Audio != nil {
				// Retrieve the file
				fmt.Println("2")
				if udp.Audio != nil {
					fmt.Println("3")
					fileConfig := tgbotapi.FileConfig{FileID: udp.Audio.FileID}
					file, err = bot.GetFile(fileConfig)
					if err != nil {
						config.Log.Error(fmt.Sprintf("Failed to retrieve audio file: %v", err))
						continue
					}
				} else {
					fmt.Println("4")
					fileConfig := tgbotapi.FileConfig{FileID: udp.Voice.FileID}
					file, err = bot.GetFile(fileConfig)
					if err != nil {
						config.Log.Error(fmt.Sprintf("Failed to retrieve voice file: %v", err))
						continue
					}
				}
				fmt.Println("5")

				// Download the file
				fileURL := file.Link(t.botToken)
				err = downloadFile(fileURL, "./audio.ogg")
				if err != nil {
					config.Log.Error(fmt.Sprintf("Failed to download audio file: %v", err))
					continue
				}

				// Send audio to ChatGPT API
				responseText, err := sendAudioToChatGPT("./audio.ogg")
				if err != nil {
					config.Log.Error(fmt.Sprintf("Failed to send audio to ChatGPT: %v", err))
					continue
				}

				fmt.Println("HOLA")

				newMessage := tgbotapi.NewMessage(update.Message.Chat.ID, responseText)
				newMessage.ParseMode = "Markdown"
				_, err = bot.Send(newMessage)
				if err != nil {
					config.Log.Errorf("Failed to send message: %v", err)
				}
			}
		}
	}

	return nil, nil
}

// downloadFileToBytes downloads the file from a given URL
func downloadFileToBytes(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func sendAudioToChatGPT(filePath string) (string, error) {
	client := openai.NewClient(os.Getenv("TOKEN"))

	// Step 1: Transcribe the audio using Whisper API with Spanish language
	audioReq := openai.AudioRequest{
		FilePath: filePath,
		Prompt:   "En español",
		Model:    "whisper-1", // Whisper model for transcription
		Language: "es",        // Specify "es" for Spanish language transcription
	}

	transcriptionResp, err := client.CreateTranscription(context.Background(), audioReq)
	if err != nil {
		fmt.Printf("Audio transcription error: %v\n", err)
		return "", err
	}

	// Step 2: Use the transcription to generate the prompt for ChatGPT
	transcriptionText := transcriptionResp.Text
	systemInput := "Eres un asistente que trabaja con nosotros desenvolviendo ideas de un proyecto de software en el que usamos SCRUM y materializamos nuestros pensamientos en tareas de Jira"
	userMessage := fmt.Sprintf("Resume esta transcripción enviada por 'alguien', en un muy breve parrafo y en caso de creer conveniente que ella se puede sacar provecho de algo en forma de tareas (solo pon el titulo)."+
		"\n"+
		"El contexto es siempre acerca de nuestra aplicación de generación de itinerarios de viaje, de itinerarios de viaje, de nuevas ideas o discusiones relativas a itinerarios, viajes, y las consiguientes aplicaciones web que estamos desarrollando."+
		"\n"+
		":\n\n%s", transcriptionText)

	chatResp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini, // or openai.GPT3Dot5Turbo
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemInput,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userMessage,
				},
			},
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return "", err
	}

	responseBody := chatResp.Choices[0].Message.Content
	fmt.Println(responseBody)

	return responseBody, nil
}
