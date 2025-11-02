package views

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/framework/gui"
	"github.com/galaco/kero/framework/gui/dialogs"
	"github.com/galaco/kero/gui/views/menu"
	"github.com/galaco/kero/messages"
)

type Menu struct {
	Console     menu.Console
	Performance *menu.Performance
	eventBus    *event.Dispatcher
	fileSystem  filesystem.FileSystem
}

// NewMenu creates a new menu view with explicit dependencies
func NewMenu(eventBus *event.Dispatcher, fileSystem filesystem.FileSystem, performance *menu.Performance) *Menu {
	return &Menu{
		eventBus:    eventBus,
		fileSystem:  fileSystem,
		Performance: performance,
	}
}

func (view *Menu) Render() {
	// StartPanel always requires a matching EndPanel, regardless of return value
	// Use NoCollapse flag to prevent the panel from being collapsed
	if gui.StartPanelV("Menu", nil, imgui.WindowFlagsNoCollapse) {
		gui.NewButton("menu_open_map", "Open map", func() {
			// Get game base path from filesystem
			gameBasePath := ""
			if view.fileSystem != nil {
				gameBasePath = filesystem.GameBasePath() // Still need to use this for now
			}
			name, err := dialogs.OpenFile("Select BSP", gameBasePath, "Valve .bsp files", "bsp")
			if err != nil {
				if err.Error() == "Cancelled" {
					return
				}
				dialogs.ErrorMessage(err)
				return
			}
			// Use deferred typed event dispatch to prevent mid-frame corruption (Phase 3)
			event.DispatchTypedDeferred(view.eventBus, event.PhasePostUpdate, messages.ChangeLevelEvent{MapName: name})
		}).Draw()
		gui.NewButton("menu_disconnect", "Disconnect", func() {
			// Use typed event dispatch (Phase 3)
			event.DispatchTyped(view.eventBus, messages.EngineDisconnectEvent{})
		}).Draw()
		gui.NewButton("menu_quit", "Quit", func() {
			// Use typed event dispatch (Phase 3)
			event.DispatchTyped(view.eventBus, messages.EngineQuitEvent{})
		}).Draw()
	}
	// ALWAYS call EndPanel after StartPanel, even if StartPanel returns false
	gui.EndPanel()

	view.Console.Render()

	// Render performance metrics
	if view.Performance != nil {
		view.Performance.Render()
	}
}
