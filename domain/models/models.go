package models

type Job struct {
	ID         string `yaml:"id"`
	CronString string `yaml:"cronString"`
	Response   []struct {
		Type    string `yaml:"type"`
		ChatID  ChatID `yaml:"chatID"`
		TopicID Topic  `yaml:"topicID"`
	} `yaml:"response"`
}

type Topic int64
type ChatID int64
type Kind int64

const (
	KindAnimation Kind = 0
	KindMedia          = 1
	KindMessage        = 2
)

type TGMessageArray []TGMessage

type TGMessage struct {
	MSG   string
	Media *[]byte
	Kind  Kind
}

func NewTGMessage(msg string, media *[]byte, kind Kind) *TGMessage {
	return &TGMessage{MSG: msg, Media: media, Kind: kind}
}
