package views

import (
	"fmt"
	"github.com/galaco/kero/framework/gui"
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
	view.text.Render()
}
