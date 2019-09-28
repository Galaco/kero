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

// Load attempts to open and unmarshall
// json configuration
func Load(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return &config, err
	}

	config.validate()

	return &config, nil
}

// validate that expected parameters with known
// boundaries or limitation fall within expectations.
func (config *Config) validate() {
	if config.Video.Width < minWidth {
		config.Video.Width = minWidth
	}

	if config.Video.Height < minHeight {
		config.Video.Height = minHeight
	}
}
