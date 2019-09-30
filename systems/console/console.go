package console

import (
	"github.com/galaco/kero/event"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/systems"
)

type Console struct {
}

func (c *Console) Register(ctx *systems.Context) {

}

func (c *Console) Update(dt float64) {

}

func (c *Console) ProcessMessage(message event.Dispatchable) {
	switch message.Type() {
	case messages.TypeConsoleMessage:
		console.PrintString(message.(*messages.ConsoleMessage).Level(), message.(*messages.ConsoleMessage).Message())
	}
}

func NewConsole() *Console {
	return &Console{}
}
