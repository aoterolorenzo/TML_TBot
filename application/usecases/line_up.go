package usecases

import (
	"TML_TBot/domain/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// URLs for website 1
var artistsURL = "https://www.tomorrowland.com/api/v2?method=LineUp.getArtists&eventid=17&format=json"
var stagesURL = "https://www.tomorrowland.com/api/v2?method=LineUp.getStages&eventid=17&format=json"

// CustomTime is a custom time type to handle the non-standard time format
type CustomTime struct {
	time.Time
}

// UnmarshalJSON customizes the unmarshalling of CustomTime
func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	str := string(b)
	// Trim the quotes
	str = str[1 : len(str)-1]

	// Parse the time
	t, err := time.Parse("2006-01-02 15:04:05-07:00", str)
	if err != nil {
		return err
	}
	ct.Time = t
	return nil
}

// Artist represents an artist with performances
type Artist struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	UID          string        `json:"uid"`
	Performances []Performance `json:"performances"`
	Image        string        `json:"image"`
	Facebook     string        `json:"facebook"`
	Twitter      string        `json:"twitter"`
	Youtube      string        `json:"youtube"`
	Soundcloud   string        `json:"soundcloud"`
	Instagram    string        `json:"instagram"`
}

// Performance represents a performance of an artist
type Performance struct {
	ID        string     `json:"id"`
	StageID   string     `json:"stage_id"`
	StartTime CustomTime `json:"start_time"`
	EndTime   CustomTime `json:"end_time"`
}

// Stage represents a stage
type Stage struct {
	ID       string `json:"id"`
	Host     string `json:"host"`
	Name     string `json:"name"`
	Priority int    `json:"priority"`
	Color    string `json:"color"`
}

// Data represents the structure of JSON data
type Data struct {
	Artists []Artist `json:"artists"`
	Stages  []Stage  `json:"stages"`
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

// MergeData merges artists and stages data into a single Data struct
func MergeData(artistsURL, stagesURL string) (Data, error) {
	var data Data
	var artists struct {
		Artists []Artist `json:"artists"`
	}
	var stages struct {
		Stages []Stage `json:"stages"`
	}

	err := FetchJSON(artistsURL, &artists)
	if err != nil {
		return data, fmt.Errorf("error fetching artists: %v", err)
	}
	err = FetchJSON(stagesURL, &stages)
	if err != nil {
		return data, fmt.Errorf("error fetching stages: %v", err)
	}

	data.Artists = artists.Artists
	data.Stages = stages.Stages

	return data, nil
}

// CompareData compares artists and performances from two sources and prints the differences
func CompareData(data1, data2 Data) string {
	// Create maps for easy lookup
	artistsMap1 := make(map[string]Artist)
	artistsMap2 := make(map[string]Artist)
	stagesMap := make(map[string]Stage)

	for _, artist := range data1.Artists {
		artistsMap1[artist.ID] = artist
	}
	for _, artist := range data2.Artists {
		artistsMap2[artist.ID] = artist
	}
	for _, stage := range data1.Stages {
		stagesMap[stage.ID] = stage
	}
	for _, stage := range data2.Stages {
		stagesMap[stage.ID] = stage
	}

	var artistsDiff strings.Builder

	// Compare artists and performances
	for id, artist1 := range artistsMap1 {
		if artist2, exists := artistsMap2[id]; exists {
			artistsDiff.WriteString(comparePerformances(artist1, artist2, stagesMap))
		} else {
			for _, perf := range artist1.Performances {
				artistsDiff.Write([]byte(fmt.Sprintf("❌ Stage %s: Eliminado <i>%s</i> (<i>%s</i>)\n", stagesMap[perf.StageID].Name, artist1.Name, perf.StartTime))) //stage, day
			}
		}
	}

	for id, artist2 := range artistsMap2 {
		if _, exists := artistsMap1[id]; !exists {
			for _, perf := range artist2.Performances {
				artistsDiff.Write([]byte(fmt.Sprintf("✅ Stage %s: Añadido <i>%s</i> (<i>%s</i>)\n", stagesMap[perf.StageID].Name, artist2.Name, perf.StartTime))) //stage, day

			}
		}
	}

	return artistsDiff.String()
}

func comparePerformances(artist1, artist2 Artist, stagesMap map[string]Stage) string {
	perfMap1 := make(map[string]Performance)
	perfMap2 := make(map[string]Performance)

	var artistsDiff strings.Builder

	for _, perf := range artist1.Performances {
		perfMap1[perf.ID] = perf
	}
	for _, perf := range artist2.Performances {
		perfMap2[perf.ID] = perf
	}

	for id, perf1 := range perfMap1 {
		if perf2, exists := perfMap2[id]; exists {
			if perf1.StageID != perf2.StageID || !perf1.StartTime.Equal(perf2.StartTime.Time) || !perf1.EndTime.Equal(perf2.EndTime.Time) {
				artistsDiff.Write([]byte(fmt.Sprintf("🔁 Stage %s: Se mueve <i>%s</i> (<i>%s</i>)\n", stagesMap[perf1.StageID].Name, artist1.Name, perf1.StartTime))) //stage, day
			}
		} else {
			artistsDiff.Write([]byte(fmt.Sprintf("❌ Stage %s: Eliminado <i>%s</i> (<i>%s</i>)\n", stagesMap[perf1.StageID].Name, artist1.Name, perf1.StartTime))) //stage, day
		}
	}

	for id, perf2 := range perfMap2 {
		if _, exists := perfMap1[id]; !exists {
			artistsDiff.Write([]byte(fmt.Sprintf("✅ Stage %s: Añadido <i>%s</i> (<i>%s</i>)\n", stagesMap[perf2.StageID].Name, artist2.Name, perf2.StartTime))) //stage, day
		}
	}

	return artistsDiff.String()
}

type TMLLineUpController struct {
	lineUp Data
}

func NewTMLLineUpController() *TMLLineUpController {

	data, err := MergeData(artistsURL, stagesURL)
	if err != nil {
		fmt.Printf("Error merging data from website 1: %v\n", err)
		return nil
	}

	return &TMLLineUpController{
		lineUp: data,
	}

}

func (l *TMLLineUpController) Run() ([]models.TGMessage, error) {
	msg := l.updateAndCompare()

	return []models.TGMessage{
		{
			MSG:   msg,
			Media: nil,
			Kind:  models.KindMessage,
		},
	}, nil
}

func (l *TMLLineUpController) updateAndCompare() string {
	data, err := MergeData(artistsURL, stagesURL)
	if err != nil {
		fmt.Printf("Error merging data from website 2: %v\n", err)
		return ""
	}

	return CompareData(l.lineUp, data)
}
