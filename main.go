package main

import (
	"TML_TBot/application/usecases"
	"TML_TBot/domain/models"
	telegram2 "TML_TBot/infrastructure/connectors"
	"fmt"
	"log"
)

func main() {
	fmt.Println("RUNNING!!!!")
	wc := usecases.WeatherController{}
	msgArray, err := wc.Run()
	if err != nil {
		log.Fatal(err)
	}
	topic := models.TopicElTiempo

	tgService := telegram2.NewTelegramService()
	for _, msg := range msgArray {
		switch msg.Kind {
		case models.KindMessage:
			err := tgService.SendMessage(msg.MSG, &topic)
			if err != nil {
				log.Fatal(err)
				return
			}
			break
		case models.KindAnimation:
			err := tgService.SendAnimation(msg.MSG, msg.Media, &topic)
			if err != nil {
				log.Fatal(err)
				return
			}
			break
		case models.KindMedia:
			err := tgService.SendMedia(msg.MSG, msg.Media, &topic)
			if err != nil {
				log.Fatal(err)
				return
			}
			break

		}
	}
}
