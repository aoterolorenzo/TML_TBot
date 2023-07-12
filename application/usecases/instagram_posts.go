package usecases

import (
	"TML_TBot/config"
	"TML_TBot/domain/models"
	"bytes"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"os"
	"reflect"

	"github.com/Davincible/goinsta/v3"
)

type InstagramPostsController struct {
	feedItems []FeedItem
	codes     CodesSet
}

type FeedItem struct {
	text  string
	image *[]byte
}

// Use a set to avoid duplicate codes
type CodesSet map[string]void

type void struct{}

const INSTAGRAM_CACHE_FILE = "./.cache/lastInstagramPosts.json"

func NewInstagramPostsController() *InstagramPostsController {
	var codes CodesSet
	err := readJSONFileToStruct(&codes, INSTAGRAM_CACHE_FILE)
	if err != nil {
		config.Log.Info("Cached posts not found. Starting from ground...")
	}

	c := &InstagramPostsController{codes: codes}
	return c
}

func (t *InstagramPostsController) Run() ([]models.TGMessage, error) {
	insta := goinsta.New(os.Getenv("ig_user"), os.Getenv("ig_pass"))

	err := insta.Login()
	if err != nil {
		fmt.Print("Error login")
	}

	initialCodes := t.codes
	items, updatedCodes, err := getRecentMedia(insta, t)

	// If there are no new posts then we don't send anything
	if reflect.DeepEqual(initialCodes, updatedCodes) {
		return []models.TGMessage{
			{MSG: "", Media: nil, Kind: models.KindMedia},
		}, nil
	}

	// Transform the items information to TGMessage objects
	messages := itemsToMessages(items)
	return messages, nil

}

func itemsToMessages(items []FeedItem) []models.TGMessage {
	messages := []models.TGMessage{}
	for _, item := range items {
		message := models.TGMessage{MSG: item.text, Media: item.image, Kind: models.KindMedia}
		messages = append(messages, message)
	}
	return messages
}

func getRecentMedia(insta *goinsta.Instagram, t *InstagramPostsController) ([]FeedItem, CodesSet, error) {

	codes := make(CodesSet)
	var member void

	acc := "tomorrowland"
	profile, err := insta.VisitProfile(acc)
	if err != nil {
		fmt.Println("Cannot visit profile", err)
	}

	feed := profile.Feed
	var items []FeedItem

	num_retrieved_feeds := 5

	for _, item := range feed.Items[0:num_retrieved_feeds] {

		image, err := downloadFile(item.Images.Versions[0].URL)
		if err != nil {
			fmt.Println("Error downloading image", err)
		}
		cbbytes := imageToByteArray(image)

		text := `<b>` + item.Caption.Text + "\n\n" + `</b>+Info: ` + fmt.Sprintf("https://www.instagram.com/p/%s/?img_index=1", item.Code)

		var current_item = FeedItem{text: text, image: &cbbytes}

		codes[item.Code] = member
		items = append(items, current_item)
	}

	t.codes = codes

	err = saveStructToJSONFile(t.codes, INSTAGRAM_CACHE_FILE)
	if err != nil {
		return nil, nil, err
	}

	return items, t.codes, err
}

func imageToByteArray(image image.Image) []byte {
	buff := new(bytes.Buffer)
	err := png.Encode(buff, image)
	if err != nil {
		fmt.Println("failed to create buffer", err)
	}
	cbbytes := buff.Bytes()
	return cbbytes
}

func downloadFile(url string) (image.Image, error) {
	response, err := http.Get(url)
	if err != nil {
		fmt.Print("Error downloading instagram image")
	}
	defer response.Body.Close()
	image, _, err := image.Decode(response.Body)

	return image, nil
}
