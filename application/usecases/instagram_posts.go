package usecases

import (
	"TML_TBot/domain/models"
	"bytes"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"os"

	"github.com/Davincible/goinsta/v3"
)

type InstagramPostsController struct {
	message string
}

type feedItem struct {
	text  string
	image *[]byte
}

func NewInstagramPostsController() *InstagramPostsController {
	var message string = "Test msg..."
	c := &InstagramPostsController{message: message}
	return c
}

func (t *InstagramPostsController) Run() ([]models.TGMessage, error) {
	insta := goinsta.New(os.Getenv("ig_user"), os.Getenv("ig_pass"))

	err := insta.Login()
	if err != nil {
		fmt.Print("Error login")
	}

	items := getRecentMedia(insta)

	return []models.TGMessage{
		{
			MSG:   items[0].text,
			Media: items[0].image,
			Kind:  models.KindMedia},
	}, nil

}

func getRecentMedia(insta *goinsta.Instagram) []feedItem {

	acc := "tomorrowland"
	profile, err := insta.VisitProfile(acc)
	if err != nil {
		fmt.Println("Cannot visit profile", err)
	}

	feed := profile.Feed
	var items []feedItem

	for _, item := range feed.Items[0:1] {

		image, err := downloadFile(item.Images.Versions[0].URL)
		if err != nil {
			fmt.Println("Error downloading image", err)
		}
		cbbytes := imageToByteArray(image)

		text := `<b>` + item.Caption.Text + `</b> +info: ` + fmt.Sprintf("https://www.instagram.com/p/%s/?img_index=1", item.Code)

		var current_item = feedItem{text: text, image: &cbbytes}
		items = append(items, current_item)
	}
	return items
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
