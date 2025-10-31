package context

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/galaco/tinygametools"
)

type ContextBindable interface {
	Handle() *tinygametools.Window
}

type Context struct {
	imguiContext *imgui.Context
	imguiBind    *imguiGlfw3
}

func (ctx *Context) Imgui() *imguiGlfw3 {
	return ctx.imguiBind
}

func (ctx *Context) Close() {
	defer ctx.imguiBind.Shutdown()
	defer imgui.DestroyContext()
}

func NewContext(window ContextBindable) *Context {
	ctx := &Context{
		imguiContext: imgui.CreateContext(),
		imguiBind:    imguiGlfw3Init(window.Handle().Handle()),
	}

	return ctx
}
