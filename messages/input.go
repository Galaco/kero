package messages

import (
	"github.com/galaco/kero/event"
	"github.com/galaco/kero/framework/input"
)

const (
	// TypeKeyPress
	TypeKeyPress   = event.Type("KeyPress")
	// TypeKeyRelease
	TypeKeyRelease = event.Type("KeyRelease")
	// TypeMouseMove  =
	TypeMouseMove  = event.Type("MouseMove")
)

// KeyPress
type KeyPress struct {
	key input.Key
}

// Type
func (msg *KeyPress) Type() event.Type {
	return TypeKeyPress
}

// Key
func (msg *KeyPress) Key() input.Key {
	return msg.key
}

// KeyPress
func NewKeyPress(key input.Key) *KeyPress {
	return &KeyPress{
		key: key,
	}
}

// KeyRelease
type KeyRelease struct {
	key input.Key
}

// Type
func (msg *KeyRelease) Type() event.Type {
	return TypeKeyRelease
}

// Key
func (msg *KeyRelease) Key() input.Key {
	return msg.key
}

// NewKeyRelease
func NewKeyRelease(key input.Key) *KeyRelease {
	return &KeyRelease{
		key: key,
	}
}

// MouseMove
type MouseMove struct {
	X, Y float64
}

// Type
func (msg *MouseMove) Type() event.Type {
	return TypeMouseMove
}

// NewMouseMove
func NewMouseMove(x, y float64) *MouseMove {
	return &MouseMove{
		X: x,
		Y: y,
	}
}
