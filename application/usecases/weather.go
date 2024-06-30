package usecases

import (
	"TML_TBot/domain/models"
	"bytes"
	"context"
	"fmt"
	cu "github.com/Davincible/chromedp-undetected"
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

func screenshot(target Target, quality int, res *[]byte, elementsToRemove string, extraQueries []string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(target.Url),
		chromedp.EmulateViewport(int64(target.ViewX), int64(target.ViewY)),
		chromedp.Sleep(6 * time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Hide specific divs
			var x string
			chromedp.EvaluateAsDevTools(`document.querySelectorAll('`+elementsToRemove+`')
					.forEach(function(el) {
						el.parentNode.removeChild(el);
					});

			`, &x).Do(ctx)

			for _, extraQuery := range extraQueries {
				chromedp.Sleep(1 * time.Second)
				chromedp.EvaluateAsDevTools(extraQuery, &x).Do(ctx)
			}
			return nil
		}),
		chromedp.Sleep(3 * time.Second),
		chromedp.FullScreenshot(res, quality),
	}
}

func (w *WeatherController) Run() ([]models.TGMessage, error) {
	msgAccuweather, err := getForecastSnapshot("Accuweather", "https://www.accuweather.com/en/be/boom/27002/july-weather/27002?year=2024",
		".monthly-calendar > :first-child, #top, .lbar-banner, .privacy-policy-banner, .fc-dialog-overlay, .fc-dialog-container, .fc-consent-root",
		325, 550, 940, 865)
	if err != nil {
		return nil, err
	}

	//msgMeteoBe, err := getForecastSnapshot("Meteo.be", "https://www.meteo.be/en/boom",
	//	".kmcc-cookie-bar--visible, .observation-pp",
	//	250, 1870, 1400, 2270)
	//if err != nil {
	//	return nil, err
	//}
	//msgMeteoBe.Pin = true

	//msgMeteoBeRain, err := getForecastSnapshot("Meteo.be (precipitaciones)", "https://www.meteo.be/en/boom",
	//	".kmcc-cookie-bar--visible, .observation-pp",
	//	230, 1850, 1350, 2260, `$('.btn-nav__list__item.style-scope.forecast-days')[2].click()`)
	//if err != nil {
	//	return nil, err
	//}

	gifmsg, err := getRandomWeatherGifMsg()
	if err != nil {
		return nil, err
	}

	if IsRainRadarTime() {
		rc := RainController{}
		msgs, err := rc.Run()
		if err == nil {
			return []models.TGMessage{msgAccuweather, msgs[0], gifmsg}, nil
		}
	}

	//return []models.TGMessage{msgAccuweather, msgMeteoBe, msgMeteoBeRain, gifmsg}, nil
	return []models.TGMessage{msgAccuweather, gifmsg}, nil
}

func getForecastSnapshot(title string, url string, elementsToRemove string, x0 int, y0 int, x1 int, y1 int, extraQueries ...string) (models.TGMessage, error) {

	// New creates a new context for use with chromedp. With this context
	// you can use chromedp as you normally would.
	ctx, cancel, err := cu.New(cu.NewConfig(
		// Remove this if you want to see a browser window.
		cu.WithHeadless(),

		// If the webelement is not found within 10 seconds, timeout.
		cu.WithTimeout(60*time.Second),
	))
	if err != nil {
		panic(err)
	}
	defer cancel()

	var target = Target{url,
		"", 1600, 1920}

	var buf []byte

	// start the browser
	if err := chromedp.Run(ctx,
		screenshot(target, 100, &buf, elementsToRemove, extraQueries)); err != nil {
		fmt.Println(err.Error())
		return models.TGMessage{}, err
	}

	// Load the captured screenshot into an image object
	img, err := imaging.Decode(bytes.NewReader(buf))
	if err != nil {
		return models.TGMessage{}, err
	}

	// Define the area to be cropped
	cropArea := image.Rect(x0, y0, x1, y1) // Example: (x1, y1, x2, y2)
	// Crop the image to the specified area
	croppedImg := imaging.Crop(img, cropArea)

	// Encode the cropped image to PNG format
	croppedBuf := new(bytes.Buffer)
	if err := imaging.Encode(croppedBuf, croppedImg, imaging.PNG); err != nil {
		return models.TGMessage{}, err
	}

	currentDate := time.Now().Format("02/01/2006")
	text := ` 
Previsión del tiempo ` + title + ` <b> a día ` + currentDate + `</b>
 
+info: ` + url
	cbbytes := croppedBuf.Bytes()

	msg1 := models.NewTGMessage(text, &cbbytes, models.KindMedia)
	return *msg1, nil
}

func getRandomWeatherGifMsg() (models.TGMessage, error) {
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(15) + 1

	filePath := "./assets/gifs/" + strconv.Itoa(randomNumber) + ".gif"

	gif2Bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return models.TGMessage{}, err
	}

	msg := models.NewTGMessage("", &gif2Bytes, models.KindAnimation)
	response := *msg
	return response, nil

}

func IsRainRadarTime() bool {
	// Define the target date
	targetDateStr := "17/07/2024"
	targetDate, err := time.Parse("02/01/2006", targetDateStr)
	if err != nil {
		return false
	}

	// Get the current date
	currentDate := time.Now()

	// Check if the current date is after the target date
	if currentDate.After(targetDate) {
		// Perform the action if the date is surpassed
		return true
	}

	return false
}
