package menu

import (
	"github.com/galaco/kero/framework/gui"
)

type Console struct {
	messages []*gui.Text
}

func (view *Console) Render() {
	if gui.StartPanel("Console") {
		for _, s := range view.messages {
			s.Render()
		}

		gui.EndPanel()
	}
}

func (view *Console) AddMessage(message string) {
	view.messages = append(view.messages, gui.NewText(message))
}
