package gui

import (
	"fmt"
	"github.com/galaco/kero/client/gui/views"
	inputMiddleware "github.com/galaco/kero/client/input"
	"github.com/galaco/kero/internal/framework/console"
	"github.com/galaco/kero/internal/framework/event"
	"github.com/galaco/kero/internal/framework/gui"
	"github.com/galaco/kero/internal/framework/gui/context"
	"github.com/galaco/kero/internal/framework/input"
	"github.com/galaco/kero/internal/framework/window"
	messages2 "github.com/galaco/kero/shared/messages"
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
		switch v := message.(type) {
		case string:
			s.menuView.Console.AddMessage(level, message.(string))
		default:
			s.menuView.Console.AddMessage(level, fmt.Sprintf("%s", v))
		}
	})
	console.DisableBufferedLogs()

	s.uiContext = context.NewContext(window.CurrentWindow())
	inputMiddleware.InputMiddleware().AddListener(messages2.TypeKeyRelease, s.onKeyRelease)
	event.Get().AddListener(messages2.TypeLoadingLevelProgress, s.onLoadingLevelProgress)

}

func (s *Gui) onKeyRelease(message interface{}) {
	key := message.(input.Key)
	if key == input.KeyEscape {
		s.shouldDisplayMenu = !s.shouldDisplayMenu
	}
}

func (s *Gui) onLoadingLevelProgress(message interface{}) {
	stage := message.(int)
	s.loadingView.UpdateProgress(stage)
	if stage == messages2.LoadingProgressStateError ||
		stage == messages2.LoadingProgressStateFinished {
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
