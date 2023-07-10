package usecases

import "TML_TBot/domain/models"

type HelloWorldController struct {
	message string
}

func NewHelloWorldController() *HelloWorldController {
	var message string = "Esta prohibido el cristal y el vidrio"
	c := &HelloWorldController{message: message}
	return c
}

func (t *HelloWorldController) Run() ([]models.TGMessage, error) {
	return []models.TGMessage{
		{MSG: t.message, Media: nil, Kind: models.KindMessage},
	}, nil
}
