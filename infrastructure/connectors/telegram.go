package connectors

import (
	"TML_TBot/domain/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"os"
)

type TelegramService struct {
	bot *tgbotapi.BotAPI
}

func NewTelegramService() *TelegramService {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	return &TelegramService{
		bot: bot,
	}
}

func (ts *TelegramService) SendMedia(msg string, media *[]byte, topic *models.Topic) error {
	imageFileBytes := tgbotapi.FileBytes{
		Name:  "img.png",
		Bytes: *media,
	}

	sendMsg := tgbotapi.NewPhotoUpload(int64(models.TMLChatID), imageFileBytes)
	sendMsg.ParseMode = "HTML"

	if msg != "" {
		sendMsg.Caption = msg
	}

	if topic != nil {
		sendMsg.ReplyToMessageID = int(*topic)
	}

	_, err := ts.bot.Send(sendMsg)
	if err != nil {
		return err
	}

	return nil
}

func (ts *TelegramService) SendMessage(msg string, topic *models.Topic) error {
	sendMsg := tgbotapi.NewMessage(int64(models.TMLChatID), msg)
	sendMsg.ParseMode = "HTML"
	sendMsg.Text = msg

	if topic != nil {
		sendMsg.ReplyToMessageID = int(*topic)
	}

	_, err := ts.bot.Send(sendMsg)
	if err != nil {
		return err
	}

	return nil
}

func (ts *TelegramService) SendAnimation(msg string, media *[]byte, topic *models.Topic) error {
	animationMsg := tgbotapi.NewAnimationUpload(int64(models.TMLChatID), tgbotapi.FileBytes{
		Name:  "random.gif",
		Bytes: *media,
	})

	animationMsg.ReplyToMessageID = int(*topic)

	if msg != "" {
		animationMsg.Caption = msg
	}

	_, err := ts.bot.Send(animationMsg)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}
