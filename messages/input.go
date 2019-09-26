package messages

import (
	"github.com/galaco/kero/event"
	"github.com/galaco/kero/framework/input"
)

const (
	TypeKeyPress   = event.Type("KeyPress")
	TypeKeyRelease = event.Type("KeyRelease")
	TypeMouseMove  = event.Type("MouseMove")
)

type KeyPress struct {
	key input.Key
}

func (msg *KeyPress) Type() event.Type {
	return TypeKeyPress
}

func (msg *KeyPress) Key() input.Key {
	return msg.key
}

func NewKeyPress(key input.Key) *KeyPress {
	return &KeyPress{
		key: key,
	}
}

type KeyRelease struct {
	key input.Key
}

func (msg *KeyRelease) Type() event.Type {
	return TypeKeyRelease
}

func (msg *KeyRelease) Key() input.Key {
	return msg.key
}

func NewKeyRelease(key input.Key) *KeyRelease {
	return &KeyRelease{
		key: key,
	}
}

type MouseMove struct {
	X, Y float64
}

func (msg *MouseMove) Type() event.Type {
	return TypeMouseMove
}

func NewMouseMove(x, y float64) *MouseMove {
	return &MouseMove{
		X: x,
		Y: y,
	}
}
