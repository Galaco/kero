package deferred

import (
	"github.com/galaco/kero/framework/graphics/adapter"
	"github.com/go-gl/gl/v4.1-core/gl"
)

func (renderer *Renderer) LightPass() {
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0)

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	renderer.directionalLightShader.Bind()
	adapter.PushInt32(renderer.directionalLightShader.GetUniform("uColorTex"), 0)
	gl.ActiveTexture(gl.TEXTURE0 + 0)
	gl.BindTexture(gl.TEXTURE_2D, renderer.gbuffer.Textures[0])

	adapter.PushInt32(renderer.directionalLightShader.GetUniform("uNormalTex"), 1)
	gl.ActiveTexture(gl.TEXTURE0 + 1)
	gl.BindTexture(gl.TEXTURE_2D, renderer.gbuffer.Textures[1])

	adapter.PushInt32(renderer.directionalLightShader.GetUniform("uPositionTex"), 2)
	gl.ActiveTexture(gl.TEXTURE0 + 2)
	gl.BindTexture(gl.TEXTURE_2D, renderer.gbuffer.Textures[2])
	gl.DrawArrays(gl.TRIANGLES, 0, 3)
}
