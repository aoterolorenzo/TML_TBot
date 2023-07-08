package config

import (
	"TML_TBot/domain/models"
	"dario.cat/mergo"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
)

import _ "embed"

//go:embed settings.yaml
var settingsFile string

var Settings = &GlobalSettings{}

type GlobalSettings struct {
	Jobs          []models.Job `yaml:"jobs"`
	ExtConfigFile string       `yaml:"externalConfigFile"`
}

func init() {
	Log.Debugf("Generating internal settings...\n")
	if err := extractSettings(settingsFile); err != nil {
		log.Panicln(err)
	}

	Log.Debugf("Retrieving settings from %s...\n", Settings.ExtConfigFile)
	yamlFile, err := ioutil.ReadFile(Settings.ExtConfigFile)
	if err != nil {
		Log.WithError(err).Debugf("File %s not available. Skipping\n", Settings.ExtConfigFile)
	} else {
		if err := extractSettings(string(yamlFile)); err != nil {
			log.Panicln(err)
		}
	}
	Log.Debugf("Settings successfully generated\n")
}

func (g *GlobalSettings) RetrieveSettingsFromFile(file string) {
	settingsFile, err := ioutil.ReadFile(file)

	if err := extractSettings(string(settingsFile)); err != nil {
		log.Panicln(err)
	}

	yamlFile, err := ioutil.ReadFile(Settings.ExtConfigFile)
	if err != nil {
	} else {
		if err := extractSettings(string(yamlFile)); err != nil {
			log.Panicln(err)
		}
	}
}

func extractSettings(content string) error {
	var setts = &GlobalSettings{}
	if err := yaml.Unmarshal([]byte(content), setts); err != nil {
		return err
	}

	if err := mergo.Merge(Settings, setts, mergo.WithOverride); err != nil {
		return err
	}

	return nil
}
