package views

import (
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/framework/gui"
	"github.com/galaco/kero/framework/gui/dialogs"
	"github.com/galaco/kero/gui/views/menu"
	"github.com/galaco/kero/messages"
)

type Menu struct {
	Console    menu.Console
	eventBus   *event.Dispatcher
	fileSystem filesystem.FileSystem
}

// NewMenu creates a new menu view with explicit dependencies
func NewMenu(eventBus *event.Dispatcher, fileSystem filesystem.FileSystem) *Menu {
	return &Menu{
		eventBus:   eventBus,
		fileSystem: fileSystem,
	}
}

func (view *Menu) Render() {
	if gui.StartPanel("Menu") {
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
			view.eventBus.Dispatch(messages.TypeChangeLevel, name)
		}).Draw()
		gui.NewButton("menu_disconnect", "Disconnect", func() {
			view.eventBus.Dispatch(messages.TypeEngineDisconnect, nil)
		}).Draw()
		gui.NewButton("menu_quit", "Quit", func() {
			view.eventBus.Dispatch(messages.TypeEngineQuit, nil)
		}).Draw()
		gui.EndPanel()
	}

	view.Console.Render()
}
