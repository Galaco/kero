package gui

import (
	"github.com/galaco/kero/internal/framework/gui/context"
	"github.com/inkyblackness/imgui-go/v4"
)

func BeginFrame(ctx *context.Context) {
	ctx.Imgui().NewFrame()
}

func EndFrame(ctx *context.Context) {
	//app.GraphicsAdapter.Viewport(0, 0, 640, 480)

	imgui.Render()
	ctx.Imgui().Render(imgui.RenderedDrawData())

	//ctx.DrawContext().Stack.Execute()
}
