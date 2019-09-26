package console

import (
	"fmt"
	"github.com/galaco/kero/event"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/systems"
)

type Console struct {
	systems.System
}

func (c *Console) ProcessMessage(message event.Dispatchable) {
	switch message.Type() {
	case messages.TypeConsoleMessage:
		console.PrintString(message.(*messages.ConsoleMessage).Level(), message.(*messages.ConsoleMessage).Message())
	default:
		console.PrintInterface(console.LevelInfo, fmt.Sprintf("%s %s", message.Type(), message))
	}
}

func NewConsole() *Console {
	return &Console{}
}
