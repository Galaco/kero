package config

import (
	"testing"
)

func TestLoad(t *testing.T) {
	_, err := Load("./../config.example.json")

	if err != nil {
		t.Error(err)
	}
}
