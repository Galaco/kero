package views

import (
	"github.com/galaco/kero/client/gui/views/menu"
	"github.com/galaco/kero/internal/framework/event"
	"github.com/galaco/kero/internal/framework/filesystem"
	"github.com/galaco/kero/internal/framework/gui"
	"github.com/galaco/kero/internal/framework/gui/dialogs"
	messages2 "github.com/galaco/kero/shared/messages"
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
			event.Get().Dispatch(messages2.TypeChangeLevel, name)
		}).Draw()
		gui.NewButton("menu_disconnect", "Disconnect", func() {
			event.Get().Dispatch(messages2.TypeEngineDisconnect, nil)
		}).Draw()
		gui.NewButton("menu_quit", "Quit", func() {
			event.Get().Dispatch(messages2.TypeEngineQuit, nil)
		}).Draw()
		gui.EndPanel()
	}

	view.Console.Render()
}
