package gui

import (
	"fmt"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/framework/gui"
	"github.com/galaco/kero/framework/gui/context"
	"github.com/galaco/kero/framework/input"
	"github.com/galaco/kero/framework/window"
	"github.com/galaco/kero/gui/views"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/middleware"
)

type Gui struct {
	eventBus        *event.Dispatcher
	fileSystem      filesystem.FileSystem
	inputMiddleware *middleware.Input
	uiContext       *context.Context

	loadingView views.Loading
	menuView    *views.Menu

	shouldDisplayMenu          bool
	shouldDisplayLoadingScreen bool
}

func (s *Gui) Initialize() {
	// Initialize menu view with dependencies
	s.menuView = views.NewMenu(s.eventBus, s.fileSystem)

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
	s.inputMiddleware.EventBus().AddListener(messages.TypeKeyRelease, s.onKeyRelease)
	s.eventBus.AddListener(messages.TypeLoadingLevelProgress, s.onLoadingLevelProgress)

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
	if stage == messages.LoadingProgressStateError ||
		stage == messages.LoadingProgressStateFinished {
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

// NewGui creates a new GUI system with explicit dependencies
func NewGui(eventBus *event.Dispatcher, fileSystem filesystem.FileSystem, inputMiddleware *middleware.Input) *Gui {
	return &Gui{
		eventBus:          eventBus,
		fileSystem:        fileSystem,
		inputMiddleware:   inputMiddleware,
		shouldDisplayMenu: true,
	}
}
