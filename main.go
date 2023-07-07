package main

import (
	"TML_TBot/controllers"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {

	/*	c := controllers.TMLLineUpController{}
		_, err := c.GetDiff()
		if err != nil {
			return
		}*/

	wc := controllers.WeatherController{}
	buff := wc.Retrieve()

	// Telegram bot token
	botToken := "6115190770:AAFh10LA0qWr93M2lLpuXC-AygF5_aF_KiI"

	// Group chat ID
	chatID := "-1001949361275"

	// Topic ID
	topicID := 2641

	// Create a new Telegram bot instance
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new file upload
	imageFileBytes := tgbotapi.FileBytes{
		Name:  "cropped_image.png",
		Bytes: buff.Bytes(),
	}

	// Calculate the combined chat ID (group chat ID + topic ID)
	fchatID, _ := strconv.ParseInt(chatID, 10, 64)

	// Create the document message
	documentMsg := tgbotapi.NewPhotoUpload(fchatID, imageFileBytes)
	// Get the current date and format it as "Sat 12 July 2023"
	currentDate := time.Now().Format("02/01/2006")
	documentMsg.Caption = `Previsión del tiempo Accuweather <b>` + currentDate + `</b>`

	documentMsg.ParseMode = "HTML"

	documentMsg.ReplyToMessageID = topicID

	// Send the document message
	_, err = bot.Send(documentMsg)
	if err != nil {
		log.Fatal(err)
	}

	sendRandomWeatherGif()

	fmt.Println("Image sent successfully!")
}

func sendRandomWeatherGif() {
	// Telegram bot token
	botToken := "6115190770:AAFh10LA0qWr93M2lLpuXC-AygF5_aF_KiI"

	// Group chat ID
	chatID := "-1001949361275"

	// Topic ID
	topicID := 2641

	// Generate a random number between 1 and 15
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(15) + 1
	// Construct the input and output file paths
	inputFilePath := "assets/brasero_gifs/" + strconv.Itoa(randomNumber) + ".mp4"
	outputFilePath := strings.Replace(inputFilePath, ".mp4", ".gif", 1)

	fmt.Println(inputFilePath)
	fmt.Println(outputFilePath)

	if _, err := os.Stat(outputFilePath); err != nil {
		// Convert the selected MP4 file to a GIF using ffmpeg
		err := convertToGIF(inputFilePath, outputFilePath)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Read the GIF file into a byte buffer
	gif2Bytes, err := ioutil.ReadFile(outputFilePath)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new Telegram bot instance
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	// Calculate the combined chat ID (group chat ID + topic ID)
	fchatID, _ := strconv.ParseInt(chatID, 10, 64)

	// Create the GIF message
	gifMsg := tgbotapi.NewAnimationUpload(fchatID, tgbotapi.FileBytes{
		Name:  "random.gif",
		Bytes: gif2Bytes,
	})

	gifMsg.ReplyToMessageID = topicID
	os.Exit(1)
	// Send the GIF message
	_, err = bot.Send(gifMsg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("GIF sent successfully!")

}

func convertToGIF(inputFilePath, outputFilePath string) error {
	// Run the ffmpeg command to convert the input MP4 file to a GIF
	cmd := exec.Command("ffmpeg", "-i", inputFilePath, outputFilePath)
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
