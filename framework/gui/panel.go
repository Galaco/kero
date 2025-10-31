package gui

import "github.com/AllenDang/cimgui-go/imgui"

func StartPanel(name string) bool {
	return imgui.Begin(name)
}

func EndPanel() {
	imgui.End()
}
