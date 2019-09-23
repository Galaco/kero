package config

import (
	"testing"
)

func TestGet(t *testing.T) {
	c := Singleton()

	if c == nil {
		t.Error("expected Config, but got nil")
	}
}

func TestLoad(t *testing.T) {
	_, err := Load("./../../config.example.json")

	if err != nil {
		t.Error(err)
	}
}
