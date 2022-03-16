package messages

import (
	"github.com/galaco/kero/internal/framework/event"
)

const (
	// TypeKeyPress
	TypeKeyPress = event.Type("input:KeyPress")
	// TypeKeyRelease
	TypeKeyRelease = event.Type("input:KeyRelease")
	// TypeMouseMove  =
	TypeMouseMove = event.Type("input:MouseMove")
)
