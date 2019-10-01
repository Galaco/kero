package views

import (
	"github.com/galaco/kero/event"
	"github.com/galaco/kero/framework/gui"
	"github.com/galaco/kero/framework/gui/dialogs"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/systems/gui/views/menu"
)

type Menu struct {
	Console menu.Console
}

func (view *Menu) Render() {
	if gui.StartPanel("Menu") {
		gui.NewButton("menu_open_map", "Open map", func() {
			name, err := dialogs.OpenFile("Valve .bsp files", "bsp")
			if err != nil {
				dialogs.ErrorMessage(err)
				return
			}
			event.Dispatch(messages.NewChangeLevel(name))
		}).Draw()
		gui.EndPanel()
	}

	view.Console.Render()
}
