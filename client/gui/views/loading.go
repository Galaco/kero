package views

import (
	"fmt"

	"github.com/galaco/kero/internal/framework/gui"
)

type Loading struct {
	state int
	text  gui.Text
}

func (view *Loading) UpdateProgress(state int) {
	view.state = state
	view.text.SetText(fmt.Sprintf("Loading: %d", state))
}

func (view *Loading) Render() {
	if gui.StartPanel("loading_map") {
		view.text.Render()
		gui.EndPanel()
	}
}
