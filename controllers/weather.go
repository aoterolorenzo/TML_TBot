package controllers

import (
	"bytes"
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"github.com/disintegration/imaging"
	"image"
	"io/ioutil"
	"log"
	"os/user"
	"time"
)

type WeatherController struct {
}

func GetCurrentUserName() string {
	currentUser, _ := user.Current()
	username := currentUser.Username
	return username
}

type Target struct {
	Url   string `json:"url"`
	Name  string `json:"name"`
	ViewX int    `json:"viewX"`
	ViewY int    `json:"viewY"`
}

func Screenshot(target Target, quality int, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(target.Url),
		chromedp.EmulateViewport(int64(target.ViewX), int64(target.ViewY)),
		chromedp.Sleep(6 * time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Hide specific divs
			var x string
			chromedp.EvaluateAsDevTools(`document.querySelectorAll('.privacy-policy-banner, .fc-dialog-overlay, .fc-dialog-container, .fc-consent-root')
					.forEach(function(el) {
						el.parentNode.removeChild(el);
					});
			`, &x).Do(ctx)
			fmt.Println(x)
			return nil
		}),
		chromedp.Sleep(3 * time.Second),
		chromedp.FullScreenshot(res, quality),
	}
}

func (w *WeatherController) Retrieve() *bytes.Buffer {

	opts := append(
		chromedp.DefaultExecAllocatorOptions[:0], // No default options to provent chrome account login problems.

		chromedp.WindowSize(1920, 1080),
		//chromedp.Headless,
		chromedp.NoSandbox,
	)

	c, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// create a new browser
	chromeCtx, cancel := chromedp.NewContext(
		c,
	)

	var target Target = Target{"https://www.accuweather.com/en/be/boom/27002/july-weather/27002", "datastudio", 966, 1308}

	var buf []byte
	// start the browser
	if err := chromedp.Run(chromeCtx,
		Screenshot(target, 100, &buf)); err != nil {
		log.Fatal(err)
	}

	// Load the captured screenshot into an image object
	img, err := imaging.Decode(bytes.NewReader(buf))
	if err != nil {
		log.Fatal(err)
	}

	// Define the area to be cropped
	cropArea := image.Rect(20, 800, 602, 1270) // Example: (x1, y1, x2, y2)
	// Crop the image to the specified area
	croppedImg := imaging.Crop(img, cropArea)

	// Encode the cropped image to PNG format
	croppedBuf := new(bytes.Buffer)
	if err := imaging.Encode(croppedBuf, croppedImg, imaging.PNG); err != nil {
		log.Fatal(err)
	}

	// Save the cropped image to a file
	file := fmt.Sprintf("cropped-report-%s.png", target.Name)
	if err := ioutil.WriteFile(file, croppedBuf.Bytes(), 0o644); err != nil {
		log.Fatal(err)
	}

	return croppedBuf
}

// fc-dialog-overlay fc-dialog-container privacy-policy-banner
