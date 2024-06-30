package usecases

import (
	"TML_TBot/domain/models"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type RainController struct {
}

const RAIN_CACHE_FILE = "./.cache/rain.mp4"
const LOGO_PATH = "./assets/img/point.png"

func (w RainController) Run() ([]models.TGMessage, error) {
	// URL of the video to download
	videoURL := "https://www.meteo.be/resources/forecasts/models/alaro40/PRECIP_en.mp4"

	// Download the video
	err := downloadFile(videoURL, RAIN_CACHE_FILE)
	if err != nil {
		fmt.Println("Error downloading video:", err)
		return nil, err
	}

	// Overlay the logo on the video and get the result as a byte slice
	modifiedVideo, err := overlayLogo(RAIN_CACHE_FILE, LOGO_PATH)
	if err != nil {
		fmt.Println("Error overlaying logo:", err)
		return nil, err
	}

	currentDate := time.Now().Format("02/01/2006 15:04:05")
	text := ` 
Radar de lluvia <b> a ` + currentDate + `</b>
 
+info: https://www.meteo.be/en/weather/forecasts/weather-model-alaro/precipitation`

	// Create a new TGMessage and assign the video data to the Media field
	message := models.TGMessage{
		MSG:   text,
		Media: &modifiedVideo,
		Kind:  0,
		Pin:   false,
	}

	//return []models.TGMessage{msgAccuweather, msgMeteoBe, msgMeteoBeRain, gifmsg}, nil
	return []models.TGMessage{message}, nil
}

// Function to download a file and save it to a local path
func downloadFile(url, filepath string) error {
	// Send a GET request to the URL
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the response body to the file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// Function to overlay a logo on a video using ffmpeg and return the result as a byte slice
func overlayLogo(videoPath, logoPath string) ([]byte, error) {
	ffmpegPath := "ffmpeg"
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "ffmpeg-output")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	// Create the temporary output file path
	outputPath := filepath.Join(tempDir, "output.mp4")

	// Define logo size, position, and duration
	logoWidth, logoHeight := 80, 70 // Resize logo to 100x100 pixels
	posX, posY := 390, 300          // Position logo at (50, 50)
	timeFrom := "00:00:06"
	timeTo := "00:00:15"
	speed := 2.0

	// Prepare the ffmpeg filter string for logo overlay and speed change
	filter := fmt.Sprintf("[1:v]scale=%d:%d[logo];[0:v][logo]overlay=%d:%d,setpts=PTS/%.2f[v]", logoWidth, logoHeight, posX, posY, speed)

	// Execute the ffmpeg command
	cmd := exec.Command(ffmpegPath, "-ss", timeFrom, "-to", timeTo, "-i", videoPath, "-i", logoPath, "-filter_complex", filter, "-map", "[v]", outputPath)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("error executing ffmpeg: %w", err)
	}

	// Read the output file into a byte slice
	output, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("error reading output file: %w", err)
	}

	return output, nil
}
