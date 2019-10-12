package console

import (
	"fmt"
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/systems"
)

var (
	fps int
	elapsed float64
)

type Console struct {
}

func (c *Console) Register(ctx *systems.Context) {

}

func (c *Console) Update(dt float64) {
	elapsed += dt
	fps++
	if elapsed >= 1 {
		elapsed -= 1
		event.Dispatch(messages.NewConsoleMessage(console.LevelInfo, fmt.Sprintf("FPS: %d", fps)))
		fps = 0
	}
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
