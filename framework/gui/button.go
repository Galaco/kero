package gui

import "github.com/inkyblackness/imgui-go/v2"

type Button struct {
	id      string
	label   string
	onPress func()
}

func (button *Button) Draw() {
	if imgui.Button(button.label) {
		button.onPress()
	}
}

func NewButton(id string, label string, onPress func()) *Button {
	return &Button{
		id:      id,
		label:   label,
		onPress: onPress,
	}
}
