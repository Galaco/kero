package gui

import (
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/gui"
	"github.com/galaco/kero/framework/gui/context"
	"github.com/galaco/kero/framework/input"
	"github.com/galaco/kero/framework/window"
	"github.com/galaco/kero/gui/views"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/middleware"
)

type Gui struct {
	uiContext *context.Context

	loadingView views.Loading
	menuView    views.Menu

	shouldDisplayMenu          bool
	shouldDisplayLoadingScreen bool
}

func (s *Gui) Initialize() {
	console.AddOutputPipe(func(level console.LogLevel, message interface{}) {
		s.menuView.Console.AddMessage(level, message.(string))
	})

	s.uiContext = context.NewContext(window.CurrentWindow())
	middleware.InputMiddleware().AddListener(messages.TypeKeyRelease, s.onKeyRelease)
	event.Get().AddListener(messages.TypeLoadingLevelProgress, s.onLoadingLevelProgress)

}

func (s *Gui) onKeyRelease(message interface{}) {
	key := message.(*messages.KeyRelease).Key()
	if key == input.KeyEscape {
		s.shouldDisplayMenu = !s.shouldDisplayMenu
	}
}

func (s *Gui) onLoadingLevelProgress(message interface{}) {
	msg := message.(*messages.LoadingLevelProgress)
	s.loadingView.UpdateProgress(msg.State())
	if msg.State() == messages.LoadingProgressStateError ||
		msg.State() == messages.LoadingProgressStateFinished {
		s.shouldDisplayLoadingScreen = false
	} else {
		s.shouldDisplayLoadingScreen = true
	}
}

func (s *Gui) Render() {
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
