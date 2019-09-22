package gameinfo

import (
	"github.com/galaco/KeyValues"
	"io"
)

var gameInfo keyvalues.KeyValue

// Get returns static gameinfo.txt keyvalues
func Get() *keyvalues.KeyValue {
	return &gameInfo
}

// Load parses a gameinfo.txt stream to a KeyValues object
func Load(stream io.Reader) (*keyvalues.KeyValue, error) {
	kvReader := keyvalues.NewReader(stream)

	kv, err := kvReader.Read()
	if err == nil {
		gameInfo = kv
	}

	return &gameInfo, err
}
