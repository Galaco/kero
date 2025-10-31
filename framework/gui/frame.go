package gui

import (
	"github.com/galaco/kero/framework/gui/context"
)

func BeginFrame(ctx *context.Context) {
	ctx.Imgui().NewFrame()
}

func EndFrame(ctx *context.Context) {
	//app.GraphicsAdapter.Viewport(0, 0, 640, 480)

	// Backend Render() handles both imgui.Render() and OpenGL rendering
	ctx.Imgui().Render()

	//ctx.DrawContext().Stack.Execute()
}
