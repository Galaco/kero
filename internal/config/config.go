package config

import (
	"encoding/json"
	"io/ioutil"
)

const minWidth = 320
const minHeight = 240

// Project configuration properties
// Engine needs to know where to locate its game data
type Config struct {
	GameDirectory string
	Video         struct {
		Width  int
		Height int
	}
}

var config Config

// Singleton returns the global config object.
func Singleton() *Config {
	return &config
}

// Load attempts to open and unmarshall
// json configuration
func Load(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return &config, err
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return &config, err
	}

	validate()

	return &config, nil
}

// validate that expected parameters with known
// boundaries or limitation fall within expectations.
func validate() {
	if config.Video.Width < minWidth {
		config.Video.Width = minWidth
	}

	if config.Video.Height < minHeight {
		config.Video.Height = minHeight
	}
}
