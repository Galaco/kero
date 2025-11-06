package gui

import "github.com/AllenDang/cimgui-go/imgui"

func StartPanel(name string) bool {
	return imgui.Begin(name)
}

func StartPanelV(name string, open *bool, flags imgui.WindowFlags) bool {
	return imgui.BeginV(name, open, flags)
}

func EndPanel() {
	imgui.End()
}
