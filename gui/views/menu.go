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
	Console menu.Console
}

func (view *Menu) Render() {
	if gui.StartPanel("Menu") {
		gui.NewButton("menu_open_map", "Open map", func() {
			name, err := dialogs.OpenFile("Select BSP", filesystem.GameBasePath(), "Valve .bsp files", "bsp")
			if err != nil {
				if err.Error() == "Cancelled" {
					return
				}
				dialogs.ErrorMessage(err)
				return
			}
			event.Get().Dispatch(messages.TypeChangeLevel, name)
		}).Draw()
		gui.NewButton("menu_disconnect", "Disconnect", func() {
			event.Get().Dispatch(messages.TypeEngineDisconnect, nil)
		}).Draw()
		gui.NewButton("menu_quit", "Quit", func() {
			event.Get().Dispatch(messages.TypeEngineQuit, nil)
		}).Draw()
		gui.EndPanel()
	}

	view.Console.Render()
}
