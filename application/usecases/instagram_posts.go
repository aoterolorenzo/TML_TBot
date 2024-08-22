package usecases

import (
	"TML_TBot/domain/models"
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
	return &InstagramPostsController{}
}

func (t *InstagramPostsController) Run() ([]models.TGMessage, error) {
	return nil, nil

}
