package usecases

import (
	"TML_TBot/domain/models"
	"bytes"
	"context"
	"github.com/chromedp/chromedp"
	"github.com/disintegration/imaging"
	"image"
	"io/ioutil"
	"math/rand"
	"strconv"
	"time"
)

type WeatherController struct {
}

type Target struct {
	Url   string `json:"url"`
	Name  string `json:"name"`
	ViewX int    `json:"viewX"`
	ViewY int    `json:"viewY"`
}

func screenshot(target Target, quality int, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(target.Url),
		chromedp.EmulateViewport(int64(target.ViewX), int64(target.ViewY)),
		chromedp.Sleep(6 * time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Hide specific divs
			var x string
			chromedp.EvaluateAsDevTools(`document.querySelectorAll('.monthly-calendar > :first-child, #top, .lbar-banner, .privacy-policy-banner, .fc-dialog-overlay, .fc-dialog-container, .fc-consent-root')
					.forEach(function(el) {
						el.parentNode.removeChild(el);
					});
			`, &x).Do(ctx)
			return nil
		}),
		chromedp.Sleep(3 * time.Second),
		chromedp.FullScreenshot(res, quality),
	}
}

func (w *WeatherController) Run() ([]models.TGMessage, error) {

	opts := append(
		chromedp.DefaultExecAllocatorOptions[:0], // No default options to prevent chrome account login problems.
		chromedp.WindowSize(1920, 1080),
		//chromedp.Headless,
		chromedp.NoSandbox,
	)

	c, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	chromeCtx, cancel := chromedp.NewContext(
		c,
	)

	var target = Target{"https://www.accuweather.com/en/be/boom/27002/july-weather/27002", "datastudio", 966, 1308}

	var buf []byte
	// start the browser
	if err := chromedp.Run(chromeCtx,
		screenshot(target, 90, &buf)); err != nil {
		return nil, err
	}

	// Load the captured screenshot into an image object
	img, err := imaging.Decode(bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}

	// Define the area to be cropped
	cropArea := image.Rect(20, 670, 600, 1160) // Example: (x1, y1, x2, y2)
	// Crop the image to the specified area
	croppedImg := imaging.Crop(img, cropArea)

	// Encode the cropped image to PNG format
	croppedBuf := new(bytes.Buffer)
	if err := imaging.Encode(croppedBuf, croppedImg, imaging.PNG); err != nil {
		return nil, err
	}

	//// Save the cropped image to a file
	//file := fmt.Sprintf("./assets/cropped--report-%s.png", target.Name)
	//if err := ioutil.WriteFile(file, croppedBuf.Bytes(), 0o644); err != nil {
	//	return nil, err
	//}

	currentDate := time.Now().Format("02/01/2006")
	text := `Previsión del tiempo Accuweather <b>` + currentDate + `</b>`
	cbbytes := croppedBuf.Bytes()

	msg1 := models.NewTGMessage(text, &cbbytes, models.KindMedia)

	gif, err := getRandomWeatherGif()
	if err != nil {
		return nil, err
	}

	msg2 := models.NewTGMessage("", gif, models.KindAnimation)
	response := []models.TGMessage{*msg1, *msg2}
	return response, nil
}

func getRandomWeatherGif() (*[]byte, error) {
	// Generate a random number between 1 and 15
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(15) + 1

	filePath := "./assets/gifs/" + strconv.Itoa(randomNumber) + ".gif"

	gif2Bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return &gif2Bytes, nil
}
