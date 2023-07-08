package usecases

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"reflect"
	"strings"
)

type TMLLineUpController struct {
	lineUp LineUp
}

type LineUp map[string]Stage
type Stage map[string]Artist
type Artist map[string]string

func (t *TMLLineUpController) Retrieve() (LineUp, error) {

	c := colly.NewCollector()
	lineUp := make(LineUp)

	c.OnHTML(".eventday", func(e *colly.HTMLElement) {
		day := e.Attr("data-eventday")
		stages := make(Stage)
		e.ForEach(".stage", func(i int, f *colly.HTMLElement) {
			stageName := f.ChildText(".stage__heading>h4")
			artists := make(Artist)
			f.ForEach(".stage__content>ul>li", func(i int, g *colly.HTMLElement) {
				artist := strings.Split(g.ChildText("a"), "\n")[0]
				timing := g.ChildText("span")
				artists[artist] = timing
			})
			stages[stageName] = artists
		})
		lineUp[day] = stages
	})

	err := c.Visit("https://www.tomorrowland.com/en/festival/line-up/stages/friday-21-july-2023")
	if err != nil {
		return nil, err
	}

	t.lineUp = lineUp

	return lineUp, err
}

func (t *TMLLineUpController) CompareLineUps(lineUp1 LineUp, lineUp2 LineUp) (string, error) {

	var diff strings.Builder

	for day, stages := range lineUp1 {
		var stagesDiff strings.Builder
		if !reflect.DeepEqual(lineUp2[day], stages) {
			for stage, artists := range stages {
				var artistsDiff strings.Builder
				if !reflect.DeepEqual(lineUp2[day][stage], artists) {
					for artist, time := range artists {
						t1, exists := lineUp2[day][stage][artist]

						if !exists {
							// Eliminado artist
							artistsDiff.Write([]byte(fmt.Sprintf("- Eliminado %s (%s)\n", artist, time))) //stage, day

						} else {
							// Existen ambos
							if t1 != time {
								artistsDiff.Write([]byte(fmt.Sprintf("- %s se mueve de las %s a las %s\n", artist, t1, time)))
							}
							delete(lineUp2[day][stage], artist)
						}

					}
					if len(lineUp2[day][stage]) >= 0 {
						for a1, t2 := range lineUp2[day][stage] {
							artistsDiff.Write([]byte(fmt.Sprintf("- Añadido %s a las (%s)\n", a1, t2))) //stage, day
						}
					}
				}
				if artistsDiff.String() != "" {
					stagesDiff.Write([]byte(fmt.Sprintf("%s Stage\n", stage)))
					stagesDiff.Write([]byte(artistsDiff.String()))
					stagesDiff.Write([]byte("\n"))
				}
			}
		}
		if stagesDiff.String() != "" {
			diff.Write([]byte(fmt.Sprintf("----- %s ----\n\n", day)))
			diff.Write([]byte(stagesDiff.String()))
		}
	}

	return diff.String(), nil
}

func (t *TMLLineUpController) GetDiff() (string, error) {

	// If first run, retrieve and return no changes
	if t.lineUp == nil {
		_, err := t.Retrieve()
		if err != nil {
			return "", err
		}
		//return "", nil
	}

	// Example changes to generate a diff with content
	delete(t.lineUp["Friday 21 July 2023"]["Atmosphere"], "Adam Beyer")
	t.lineUp["Friday 21 July 2023"]["Cage"]["Patrick Mason"] = "26:00"

	initialLineUp := t.lineUp
	updatedLineUp, err := t.Retrieve()
	if err != nil {
		return "", err
	}

	// Example changes to generate a diff with content
	delete(t.lineUp["Friday 21 July 2023"]["Terra Solis"], "Mosoo")

	diff, err := t.CompareLineUps(initialLineUp, updatedLineUp)
	if err != nil {
		return "", err
	}

	fmt.Println(diff)

	return diff, nil
}
