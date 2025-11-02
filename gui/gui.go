package gui

import (
	"fmt"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/framework/gui"
	"github.com/galaco/kero/framework/gui/context"
	"github.com/galaco/kero/framework/input"
	"github.com/galaco/kero/framework/metrics"
	"github.com/galaco/kero/framework/window"
	"github.com/galaco/kero/gui/views"
	"github.com/galaco/kero/gui/views/menu"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/middleware"
)

// ISceneManager interface for scene management operations
type ISceneManager interface {
	CancelLoading()
}

type Gui struct {
	eventBus         *event.Dispatcher
	fileSystem       filesystem.FileSystem
	inputMiddleware  *middleware.Input
	metricsCollector *metrics.Collector
	uiContext        *context.Context
	sceneManager     ISceneManager

	loadingView *views.Loading
	menuView    *views.Menu

	shouldDisplayMenu          bool
	shouldDisplayLoadingScreen bool
}

func (s *Gui) Initialize() {
	// Initialize menu view with dependencies
	performanceView := menu.NewPerformance(s.metricsCollector)
	s.menuView = views.NewMenu(s.eventBus, s.fileSystem, performanceView)

	// Initialize loading view with cancel callback
	s.loadingView = views.NewLoading(func() {
		// Cancel button callback - trigger scene to cancel loading
		if s.sceneManager != nil {
			s.sceneManager.CancelLoading()
		}
	})

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

	// Register typed event listeners (Phase 3)
	event.RegisterTypedEvent(s.inputMiddleware.EventBus(), s.onKeyReleaseTyped)
	event.RegisterTypedEvent(s.eventBus, s.onLoadingLevelProgressTyped)

}

func (s *Gui) onKeyReleaseTyped(e messages.KeyReleaseEvent) {
	if e.Key == input.KeyEscape {
		s.shouldDisplayMenu = !s.shouldDisplayMenu
	}
}

func (s *Gui) onLoadingLevelProgressTyped(e messages.LoadingLevelProgressEvent) {
	s.loadingView.UpdateProgress(e.State)
	if e.State == messages.LoadingProgressStateError ||
		e.State == messages.LoadingProgressStateFinished {
		s.shouldDisplayLoadingScreen = false
	} else {
		s.shouldDisplayLoadingScreen = true
	}
}

func (s *Gui) Render(dt float32) {
	gui.BeginFrame(s.uiContext)

	// Apply performance ConVars
	s.applyPerformanceConVars()

	// Do rendering
	if s.shouldDisplayLoadingScreen {
		// Update loading animation
		s.loadingView.Update(dt)
		s.loadingView.Render()
	} else {
		if s.shouldDisplayMenu {
			s.menuView.Render()
		}
	}

	gui.EndFrame(s.uiContext)
}

// applyPerformanceConVars checks and applies performance-related ConVars
func (s *Gui) applyPerformanceConVars() {
	if s.metricsCollector == nil || s.menuView == nil || s.menuView.Performance == nil {
		return
	}

	// Check if metrics are enabled
	enabled := console.GetConvarBoolean("r_showperf")
	s.metricsCollector.SetEnabled(enabled)

	// Apply history size
	historySize := console.GetConvarInt("r_perfhistory")
	if historySize > 0 {
		s.metricsCollector.SetHistorySize(historySize)
	}

	// Apply graph dimensions to performance view
	height := console.GetConvarInt("r_perfgraphheight")
	if height > 0 {
		s.menuView.Performance.SetGraphHeight(float32(height))
	}

	width := console.GetConvarInt("r_perfgraphwidth")
	if width > 0 {
		s.menuView.Performance.SetGraphWidth(float32(width))
	}
}

// NewGui creates a new GUI system with explicit dependencies
func NewGui(eventBus *event.Dispatcher, fileSystem filesystem.FileSystem, inputMiddleware *middleware.Input, metricsCollector *metrics.Collector, sceneManager ISceneManager) *Gui {
	return &Gui{
		eventBus:          eventBus,
		fileSystem:        fileSystem,
		inputMiddleware:   inputMiddleware,
		metricsCollector:  metricsCollector,
		sceneManager:      sceneManager,
		shouldDisplayMenu: true,
	}
}
