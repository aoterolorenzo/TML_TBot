package usecases

import (
	"TML_TBot/domain/models"
	"encoding/json"
	"fmt"
	"github.com/k0kubun/pp/v3"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// URLs for website 1
var W1Url = "https://artist-lineup-cdn.tomorrowland.com/TLBE24-W1-211903bb-da4c-445d-a1b3-6b17479a9fab.json"
var W2Url = "https://artist-lineup-cdn.tomorrowland.com/TLBE24-W2-211903bb-da4c-445d-a1b3-6b17479a9fab.json"

const LINEUP_CACHE_FILE = "./.cache/lineUp.json"

// CustomTime is a custom time type to handle the non-standard time format
type CustomTime struct {
	time.Time
}

// UnmarshalJSON parses a time string into a CustomTime
func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	// Trim the quotes around the time string
	timeStr := strings.Trim(string(b), "\"")

	const layout = `2006-01-02 15:04:05-07:00`
	t, err := time.Parse(layout, timeStr)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05-07:00", timeStr)
		if err != nil {
			pp.Println(err)
		}
	}

	ct.Time = t

	return nil
}

func (ct CustomTime) To12HourFormat() string {
	return ct.Time.Format("03:04pm")
}

// Artist represents an artist with performances
type Artist struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Image     string `json:"image"`
	Instagram string `json:"instagram"`
	Spotify   string `json:"spotify"`
}

// Stage represents a stage
type Stage struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Performance represents a performance of an artist
type Performance struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Artists   []Artist   `json:"artists"`
	Stage     Stage      `json:"stage"`
	Date      string     `json:"date"`
	Day       string     `json:"day"`
	StartTime CustomTime `json:"startTime"`
	EndTime   CustomTime `json:"endTime"`
}

// Data represents the structure of JSON data
type Data struct {
	Performances []Performance `json:"performances"`
}

// FetchJSON fetches JSON data from a URL
func FetchJSON(url string, target interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, target)
}

// RetrieveData retrieves data from a URL into a single Data struct
func RetrieveData(url string) (Data, error) {
	var data Data
	err := FetchJSON(url, &data)
	if err != nil {
		return data, fmt.Errorf("error fetching data: %v", err)
	}
	return data, nil
}

// CompareData compares performances from two sources and prints the differences
func CompareData(data1, data2 Data) string {
	// Create maps for easy lookup
	perfMap1 := make(map[string]Performance)
	perfMap2 := make(map[string]Performance)

	for _, perf := range data1.Performances {
		perfMap1[perf.ID] = perf
	}
	for _, perf := range data2.Performances {
		perfMap2[perf.ID] = perf
	}

	var performancesDiff strings.Builder

	// Compare performances
	for id, perf1 := range perfMap1 {
		if perf2, exists := perfMap2[id]; exists {
			if perf1.Stage.ID != perf2.Stage.ID || !perf1.StartTime.Equal(perf2.StartTime.Time) || !perf1.EndTime.Equal(perf2.EndTime.Time) {
				performancesDiff.WriteString(fmt.Sprintf("🔁 <b>%s</b>: Se mueve <b>%s</b> de %s/%s/%s a %s/%s/%s\n", perf1.Stage.Name, perf1.Name,
					whichWeekend(perf1.StartTime), perf1.Day, perf1.StartTime.To12HourFormat(),
					whichWeekend(perf2.StartTime), perf2.Day, perf2.StartTime.To12HourFormat()))
			}
		} else {

			performancesDiff.WriteString(fmt.Sprintf("❌ <b>%s</b>: Eliminado <b>%s</b> (<i>%s/%s/%s</i>)\n", perf1.Stage.Name, perf1.Name, whichWeekend(perf1.StartTime), perf1.Day, perf1.StartTime.To12HourFormat()))
		}
	}

	for id, perf2 := range perfMap2 {
		if _, exists := perfMap1[id]; !exists {
			performancesDiff.WriteString(fmt.Sprintf("✅ <b>%s</b>: Añadido <b>%s</b> (<i>%s/%s/%s</i>)\n", perf2.Stage.Name, perf2.Name, whichWeekend(perf2.StartTime), perf2.Day, perf2.StartTime.To12HourFormat()))
		}
	}

	return performancesDiff.String()
}

type TMLLineUpController struct {
	Performances Data
}

func NewTMLLineUpController() *TMLLineUpController {
	var performances Data
	var fileData Data
	err := ReadStructFromJSONFile(LINEUP_CACHE_FILE, &fileData)
	if err != nil {
		fmt.Println(err)
	}

	if len(fileData.Performances) == 0 {
		data, err := RetrieveData(W1Url)
		performances.Performances = data.Performances
		if err != nil {
			fmt.Printf("Error merging data from Performances: %v\n", err)
			return nil
		}

		data2, err := RetrieveData(W2Url)
		performances.Performances = append(performances.Performances, data2.Performances...)
		if err != nil {
			fmt.Printf("Error merging data from w2: %v\n", err)
			return nil
		}

	} else {
		performances = Data{fileData.Performances}
	}

	return &TMLLineUpController{
		Performances: performances,
	}
}

func (l *TMLLineUpController) Run() ([]models.TGMessage, error) {
	msg := l.updateAndCompare()
	fmt.Println(msg)
	return []models.TGMessage{
		{
			MSG:   msg,
			Media: nil,
			Kind:  models.KindMessage,
		},
	}, nil
}

func (l *TMLLineUpController) updateAndCompare() string {
	dataW1, err := RetrieveData(W1Url)
	if err != nil {
		fmt.Printf("Error merging data from website 1: %v\n", err)
		return ""
	}

	dataW2, err := RetrieveData(W2Url)
	if err != nil {
		fmt.Printf("Error merging data from website 2: %v\n", err)
		return ""
	}

	performances := append(dataW1.Performances, dataW2.Performances...)

	response := CompareData(l.Performances, Data{Performances: performances})

	err = WriteStructToJSONFile(LINEUP_CACHE_FILE, Data{Performances: performances})
	if err != nil {
		return err.Error()
	}
	return response
}

// WriteStructToJSONFile writes any struct to a JSON file
func WriteStructToJSONFile(filename string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	err = ioutil.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ReadStructFromJSONFile reads data from a JSON file into the provided struct
func ReadStructFromJSONFile(filename string, data interface{}) error {
	fileData, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	err = json.Unmarshal(fileData, data)
	if err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}

func whichWeekend(time CustomTime) string {
	if time.Day() < 25 {
		return "W1"
	}
	return "W2"
}
