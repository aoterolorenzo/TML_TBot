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
	codes CodesSet
}

type FeedItem struct {
	text  string
	image *[]byte
}

// Use a set to avoid duplicate codes
type CodesSet map[string]any

const INSTAGRAM_CACHE_FILE = "./.cache/lastInstagramPosts.json"
const INSTAGRAM_LOGIN_CACHE_FILE = "./.cache/.goinsta"

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

	var insta *goinsta.Instagram
	insta, err := goinsta.Import(INSTAGRAM_LOGIN_CACHE_FILE)
	if err != nil {
		config.Log.Info("Cookies not found. Login in with user/password...")
		insta = goinsta.New(os.Getenv("ig_user"), os.Getenv("ig_pass"))
		err := insta.Login()
		if err != nil {
			config.Log.Fatal("Error login.", err)
		}
		defer func(ins *goinsta.Instagram, path string) {
			err := ins.Export(path)
			if err != nil {
			}
		}(insta, INSTAGRAM_LOGIN_CACHE_FILE)
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

	missingCodes := make(CodesSet)

	acc := "tomorrowland"
	profile, err := insta.VisitProfile(acc)
	if err != nil {
		config.Log.Fatal("Cannot visit instagram profile", err)
	}

	feed := profile.Feed
	var items []FeedItem

	num_retrieved_feeds := 5

	for _, item := range feed.Items[0:num_retrieved_feeds] {
		if _, ok := t.codes[item.Code]; !ok {
			image, err := downloadFile(item.Images.Versions[0].URL)
			if err != nil {
				config.Log.Fatal("Error downloading image", err)
			}
			cbbytes := imageToByteArray(image)

			text := `<b>` + item.Caption.Text + "\n\n" + `</b>+Info: ` + fmt.Sprintf("https://www.instagram.com/p/%s/?img_index=1", item.Code)

			var current_item = FeedItem{text: text, image: &cbbytes}

			missingCodes[item.Code] = nil
			items = append(items, current_item)
		}
	}

	// Merge missing codes with existing codes
	t.codes = mergeCodes(t.codes, missingCodes)

	if len(missingCodes) > 0 {
		err = saveStructToJSONFile(t.codes, INSTAGRAM_CACHE_FILE)
		if err != nil {
			return nil, nil, err
		}
	}

	return items, t.codes, err
}

func mergeCodes(existingCodes, missingCodes CodesSet) CodesSet {
	mergedCodes := make(CodesSet)
	for code := range existingCodes {
		mergedCodes[code] = nil
	}
	for code := range missingCodes {
		mergedCodes[code] = nil
	}
	return mergedCodes
}

func imageToByteArray(image image.Image) []byte {
	buff := new(bytes.Buffer)
	err := png.Encode(buff, image)
	if err != nil {
		config.Log.Fatal("failed to create buffer", err)
	}
	cbbytes := buff.Bytes()
	return cbbytes
}

func downloadFile(url string) (image.Image, error) {
	response, err := http.Get(url)
	if err != nil {
		config.Log.Fatal("Error downloading instagram image")
	}
	defer response.Body.Close()
	image, _, err := image.Decode(response.Body)

	return image, nil
}
