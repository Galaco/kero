package gui

import "github.com/inkyblackness/imgui-go/v2"

func StartPanel(name string) bool {
	return imgui.Begin(name)
}

func EndPanel() {
	imgui.End()
}
