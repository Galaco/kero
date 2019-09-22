package gameinfo

import (
	"github.com/galaco/KeyValues"
	"os"
)

// LoadConfig loads a gameinfo.txt source engine file
func LoadConfig(gameDirectory string) (*keyvalues.KeyValue, error) {
	// Load gameinfo.txt
	gameInfoFile, err := os.Open(gameDirectory + "/gameinfo.txt")
	if err != nil {
		return nil, err
	}
	return Load(gameInfoFile)
}
