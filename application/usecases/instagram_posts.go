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

	// All currently fetched timeline posts
	//posts := insta.Timeline.Items
	//image, err := downloadFile(posts[0].Images.Versions[0].URL)

	//cbbytes := imageToByteArray(image)

	getRecentMedia(insta)

	return []models.TGMessage{
		{
			MSG:   "",
			Media: nil,
			Kind:  models.KindMedia},
	}, nil
	/*
		return []models.TGMessage{
			{
				MSG:   posts[0].Caption.Text,
				Media: &cbbytes,
				Kind:  models.KindMedia},
		}, nil
	*/
}

func getRecentMedia(insta *goinsta.Instagram) {
	acc := "tomorrowland"
	profile, err := insta.VisitProfile(acc)
	if err != nil {
		fmt.Println("Cannot visit profile", err)
	}

	user := profile.User
	fmt.Printf(
		"%s has %d followers, %d posts, and %d IGTV vids\n",
		acc, user.FollowerCount, user.MediaCount, user.IGTVCount,
	)

	feed := profile.Feed
	fmt.Printf("%d posts fetched, more available = %v\n", len(feed.Items), feed.MoreAvailable)

	for _, item := range feed.Items {
		fmt.Printf("Caption: %v\n", item.Caption.Text)
		fmt.Printf("Images: %v\n", item.Images.Versions)
	}

	stories := profile.Stories
	fmt.Printf("%s currently has %d story posts\n", acc, stories.Reel.MediaCount)
}

func imageToByteArray(image image.Image) []byte {
	// create buffer
	buff := new(bytes.Buffer)

	// encode image to buffer
	err := png.Encode(buff, image)
	if err != nil {
		fmt.Println("failed to create buffer", err)
	}

	cbbytes := buff.Bytes()
	return cbbytes
}

func downloadFile(url string) (image.Image, error) {
	//Get the response bytes from the url
	response, err := http.Get(url)
	if err != nil {
		fmt.Print("Error downloading instagram image")
	}
	defer response.Body.Close()
	image, _, err := image.Decode(response.Body)

	return image, nil
}
