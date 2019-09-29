package views

import (
	"github.com/galaco/kero/event"
	"github.com/galaco/kero/framework/gui"
	"github.com/galaco/kero/framework/gui/dialogs"
	"github.com/galaco/kero/messages"
)

type Menu struct {

}

func (view *Menu) Render() {
	gui.NewButton("1", "Open map", func() {
		name, err := dialogs.OpenFile("Valve .bsp files", "bsp")
		if err != nil {
			dialogs.ErrorMessage(err)
			return
		}
		event.Dispatch(messages.NewChangeLevel(name))
	}).Draw()
}
