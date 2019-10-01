package gui

import "github.com/inkyblackness/imgui-go"

func StartPanel(name string) bool {
	return imgui.Begin(name)
}

func EndPanel() {
	imgui.End()
}
