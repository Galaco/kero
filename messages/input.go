package messages

import (
	"github.com/galaco/kero/event/message"
	"github.com/galaco/kero/framework/input"
)

const (
	TypeKeyPress   = message.Type("KeyPress")
	TypeKeyRelease = message.Type("KeyRelease")
	TypeMouseMove   = message.Type("MouseMove")
)

type KeyPress struct {
	key input.Key
}

func (msg *KeyPress) Type() message.Type {
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

func (msg *KeyRelease) Type() message.Type {
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

func (msg *MouseMove) Type() message.Type {
	return TypeMouseMove
}

func NewMouseMove(x, y float64) *MouseMove {
	return &MouseMove{
		X: x,
		Y: y,
	}
}
