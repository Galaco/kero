package gui

import "github.com/inkyblackness/imgui-go/v4"

type Text struct {
	value string
}

func (text *Text) Render() {
	imgui.Text(text.value)
}

func (text *Text) SetText(value string) {
	text.value = value
}

func NewText(value string) *Text {
	return &Text{value: value}
}
