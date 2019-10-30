package deferred

import (
	"github.com/galaco/kero/framework/graphics/adapter"
	"github.com/go-gl/gl/v4.1-core/gl"
)

func (renderer *Renderer) LightPass() {
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0)

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	renderer.gbuffer.BindReadOnly()

	halfWidth := renderer.width / 2.0
	halfHeight := renderer.height / 2.0

	renderer.gbuffer.SetReadBuffer(adapter.GBufferTextureTypePosition)
	gl.BlitFramebuffer(0, 0, renderer.width, renderer.height,
		0, 0, halfWidth, halfHeight, gl.COLOR_BUFFER_BIT, gl.LINEAR)

	renderer.gbuffer.SetReadBuffer(adapter.GBufferTextureTypeDiffuse)
	gl.BlitFramebuffer(0, 0, renderer.width, renderer.height,
		0, halfHeight, halfWidth, renderer.height, gl.COLOR_BUFFER_BIT, gl.LINEAR)

	renderer.gbuffer.SetReadBuffer(adapter.GBufferTextureTypeNormal)
	gl.BlitFramebuffer(0, 0, renderer.width, renderer.height,
		halfWidth, halfHeight, renderer.width, renderer.height, gl.COLOR_BUFFER_BIT, gl.LINEAR)

	renderer.gbuffer.SetReadBuffer(adapter.GBufferTextureTypeTextureCoordinate)
	gl.BlitFramebuffer(0, 0, renderer.width, renderer.height,
		halfWidth, 0, renderer.width, halfHeight, gl.COLOR_BUFFER_BIT, gl.LINEAR)
}
