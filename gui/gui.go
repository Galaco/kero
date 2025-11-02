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
	"github.com/galaco/kero/gui/layer"
	"github.com/galaco/kero/gui/views"
	"github.com/galaco/kero/gui/views/menu"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/middleware"
)

// ISceneManager interface for scene management operations
type ISceneManager interface {
	CancelLoading()
	IsLevelLoaded() bool
}

type Gui struct {
	eventBus         *event.Dispatcher
	fileSystem       filesystem.FileSystem
	inputMiddleware  *middleware.Input
	metricsCollector *metrics.Collector
	uiContext        *context.Context
	sceneManager     ISceneManager

	// Layer system
	layers map[layer.Layer]*layer.Config

	// Views organized by layer
	hudView     *views.HUD
	menuView    *views.Menu
	loadingView *views.Loading
}

func (s *Gui) Initialize() {
	// Initialize layer configurations
	s.layers = make(map[layer.Layer]*layer.Config)
	s.layers[layer.LayerHUD] = layer.NewConfig(layer.LayerHUD)
	s.layers[layer.LayerMenu] = layer.NewConfig(layer.LayerMenu)
	s.layers[layer.LayerModal] = layer.NewConfig(layer.LayerModal)

	// HUD layer is always visible
	s.layers[layer.LayerHUD].Visible = true

	// Menu layer starts visible (will be toggled by ESC key)
	s.layers[layer.LayerMenu].Visible = true

	// Modal layer starts hidden
	s.layers[layer.LayerModal].Visible = false

	// Initialize HUD view (Layer 0: always visible)
	performanceView := menu.NewPerformance(s.metricsCollector)
	s.hudView = views.NewHUD(performanceView)

	// Initialize menu view (Layer 100: toggle-able)
	s.menuView = views.NewMenu(s.eventBus, s.fileSystem)

	// Initialize loading view (Layer 200: modal)
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
		// Only allow closing the menu if a level is loaded
		// Menu can always be opened, but closing requires a loaded level
		menuLayer := s.layers[layer.LayerMenu]
		if !menuLayer.Visible {
			// Opening the menu - always allowed
			menuLayer.Visible = true
		} else {
			// Trying to close the menu - only allowed if level is loaded
			if s.sceneManager != nil && s.sceneManager.IsLevelLoaded() {
				menuLayer.Visible = false
			}
		}
	}
}

func (s *Gui) onLoadingLevelProgressTyped(e messages.LoadingLevelProgressEvent) {
	s.loadingView.UpdateProgress(e.State)
	modalLayer := s.layers[layer.LayerModal]
	if e.State == messages.LoadingProgressStateError ||
		e.State == messages.LoadingProgressStateFinished {
		modalLayer.Visible = false
	} else {
		modalLayer.Visible = true
	}
}

func (s *Gui) Render(dt float32) {
	gui.BeginFrame(s.uiContext)

	// Apply performance ConVars
	s.applyPerformanceConVars()

	// Check if modal layer is active (blocks lower layers)
	modalActive := s.layers[layer.LayerModal].Visible

	// Render layers in order (lowest to highest z-order)
	// Layer 0: HUD (always visible, never blocked)
	s.renderLayer(layer.LayerHUD, dt)

	// Layer 100: Menu (toggle-able, hidden when modal is active)
	if !modalActive {
		s.renderLayer(layer.LayerMenu, dt)
	}

	// Layer 200: Modal (loading screen, dialogs - blocks all lower interactive layers)
	s.renderLayer(layer.LayerModal, dt)

	gui.EndFrame(s.uiContext)
}

// renderLayer renders a specific layer if it's visible
func (s *Gui) renderLayer(l layer.Layer, dt float32) {
	layerConfig := s.layers[l]
	if !layerConfig.Visible {
		return
	}

	switch l {
	case layer.LayerHUD:
		if s.hudView != nil {
			s.hudView.Render(dt)
		}
	case layer.LayerMenu:
		if s.menuView != nil {
			s.menuView.Render(dt)
		}
	case layer.LayerModal:
		if s.loadingView != nil {
			s.loadingView.Render(dt)
		}
	}
}

// applyPerformanceConVars checks and applies performance-related ConVars
func (s *Gui) applyPerformanceConVars() {
	if s.metricsCollector == nil || s.hudView == nil || s.hudView.Performance == nil {
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
		s.hudView.Performance.SetGraphHeight(float32(height))
	}

	width := console.GetConvarInt("r_perfgraphwidth")
	if width > 0 {
		s.hudView.Performance.SetGraphWidth(float32(width))
	}
}

// NewGui creates a new GUI system with explicit dependencies
func NewGui(eventBus *event.Dispatcher, fileSystem filesystem.FileSystem, inputMiddleware *middleware.Input, metricsCollector *metrics.Collector, sceneManager ISceneManager) *Gui {
	return &Gui{
		eventBus:         eventBus,
		fileSystem:       fileSystem,
		inputMiddleware:  inputMiddleware,
		metricsCollector: metricsCollector,
		sceneManager:     sceneManager,
	}
}
