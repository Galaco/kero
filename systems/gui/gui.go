package gui

import (
	"github.com/galaco/kero/event"
	"github.com/galaco/kero/event/message"
	"github.com/galaco/kero/framework/gui"
	"github.com/galaco/kero/framework/gui/context"
	"github.com/galaco/kero/framework/gui/dialogs"
	"github.com/galaco/kero/framework/window"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/systems"
)

type Gui struct {
	systems.System
	uiContext *context.Context
}

func (s *Gui) Register() {
	s.uiContext = context.NewContext(window.CurrentWindow())
}

func (s *Gui) ProcessMessage(message message.Dispatchable) {

}

func (s *Gui) Update(dt float64) {
	gui.BeginFrame(s.uiContext)

	// Do rendering
	gui.NewButton("1", "Open map", func() {
		name, err := dialogs.OpenFile("Valve .bsp files", "bsp")
		if err != nil {
			dialogs.ErrorMessage(err)
			return
		}
		event.Singleton().Dispatch(messages.NewChangeLevel(name))
	}).Draw()

	gui.EndFrame(s.uiContext)
}

func NewGui() *Gui {
	return &Gui{}
}
