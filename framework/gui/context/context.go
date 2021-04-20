package context

import (
	"github.com/galaco/tinygametools"
	"github.com/inkyblackness/imgui-go/v4"
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
	defer ctx.imguiContext.Destroy()
	defer ctx.imguiBind.Shutdown()
}

func NewContext(window ContextBindable) *Context {
	ctx := &Context{
		imguiContext: imgui.CreateContext(nil),
		imguiBind:    imguiGlfw3Init(window.Handle().Handle()),
	}

	return ctx
}
