package gui

import (
	"github.com/galaco/kero/event"
	"github.com/galaco/kero/framework/gui"
	"github.com/galaco/kero/framework/gui/context"
	"github.com/galaco/kero/framework/input"
	"github.com/galaco/kero/framework/window"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/systems"
	"github.com/galaco/kero/systems/gui/views"
)

type Gui struct {
	uiContext *context.Context

	loadingView views.Loading
	menuView    views.Menu

	shouldDisplayMenu          bool
	shouldDisplayLoadingScreen bool
}

func (s *Gui) Register(ctx *systems.Context) {
	s.uiContext = context.NewContext(window.CurrentWindow())
}

func (s *Gui) ProcessMessage(message event.Dispatchable) {
	switch message.Type() {
	case messages.TypeKeyRelease:
		key := message.(*messages.KeyRelease).Key()
		if key == input.KeyEscape {
			s.shouldDisplayMenu = !s.shouldDisplayMenu
		}
	case messages.TypeLoadingLevelProgress:
		msg := message.(*messages.LoadingLevelProgress)
		s.loadingView.UpdateProgress(msg.State())
		if msg.State() == messages.LoadingProgressStateError ||
			msg.State() == messages.LoadingProgressStateFinished {
			s.shouldDisplayLoadingScreen = false
		} else {
			s.shouldDisplayLoadingScreen = true
		}
	case messages.TypeConsoleMessage:
		s.menuView.Console.AddMessage(message.(*messages.ConsoleMessage).Level(), message.(*messages.ConsoleMessage).Message())
	}
}

func (s *Gui) Update(dt float64) {
	gui.BeginFrame(s.uiContext)

	// Do rendering
	if s.shouldDisplayLoadingScreen {
		s.loadingView.Render()
	} else {
		if s.shouldDisplayMenu {
			s.menuView.Render()
		}
	}

	gui.EndFrame(s.uiContext)
}

func NewGui() *Gui {
	return &Gui{
		shouldDisplayMenu: true,
	}
}
