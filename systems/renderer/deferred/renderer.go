package deferred

import (
	"errors"
	"github.com/galaco/gosigl"
	"github.com/galaco/kero/framework/graphics"
	graphics3d "github.com/galaco/kero/framework/graphics/3d"
	"github.com/galaco/kero/framework/graphics/adapter"
	"github.com/go-gl/gl/v4.1-core/gl"
)

type Renderer struct {
	gbuffer       adapter.GBuffer
	width, height int32

	geometryShader         *graphics.Shader
	directionalLightShader *graphics.Shader
	activeShader           *graphics.Shader
}

func (renderer *Renderer) Init(width, height int) error {
	renderer.width = int32(width)
	renderer.height = int32(height)
	success := renderer.gbuffer.Initialize(width, height)
	if !success {
		return errors.New("failed to initialize gbuffer")
	}

	renderer.geometryShader = graphics.NewShader()
	if err := renderer.geometryShader.Add(gosigl.VertexShader, GeometryPassVertex); err != nil {
		return err
	}
	if err := renderer.geometryShader.Add(gosigl.FragmentShader, GeometryPassFragment); err != nil {
		return err
	}
	renderer.geometryShader.Finish()
	renderer.bindShader(renderer.geometryShader)
	adapter.PushInt32(renderer.geometryShader.GetUniform("albedoSampler"), 0)

	renderer.directionalLightShader = graphics.NewShader()
	if err := renderer.directionalLightShader.Add(gosigl.VertexShader, DirectionalLightPassVertex); err != nil {
		return err
	}
	if err := renderer.directionalLightShader.Add(gosigl.FragmentShader, DirectionalLightPassFragment); err != nil {
		return err
	}
	renderer.directionalLightShader.Finish()
	return nil
}

// ActiveShader
func (renderer *Renderer) ActiveShader() *graphics.Shader {
	return renderer.activeShader
}

func (renderer *Renderer) bindShader(shader *graphics.Shader) {
	shader.Bind()
	renderer.activeShader = shader
}

func (renderer *Renderer) GeometryPass(camera *graphics3d.Camera) {
	renderer.gbuffer.BindReadWrite()
	adapter.ClearColor(0,0,0.3,1)
	adapter.ClearAll()

	renderer.bindShader(renderer.geometryShader)

	adapter.PushMat4(renderer.geometryShader.GetUniform("projection"), 1, false, camera.ProjectionMatrix())
	adapter.PushMat4(renderer.geometryShader.GetUniform("view"), 1, false, camera.ViewMatrix())
	adapter.PushMat4(renderer.geometryShader.GetUniform("model"), 1, false, camera.ModelMatrix())
	adapter.PushInt32(renderer.geometryShader.GetUniform("albedoSampler"), 0)
}

func (renderer *Renderer) DirectionalLightPass() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	renderer.bindShader(renderer.directionalLightShader)

	adapter.PushInt32(renderer.directionalLightShader.GetUniform("uPositionTex"), 0)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, renderer.gbuffer.PositionBuffer)

	adapter.PushInt32(renderer.directionalLightShader.GetUniform("uNormalTex"), 1)
	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, renderer.gbuffer.NormalBuffer)

	adapter.PushInt32(renderer.directionalLightShader.GetUniform("uColorTex"), 2)
	gl.ActiveTexture(gl.TEXTURE2)
	gl.BindTexture(gl.TEXTURE_2D, renderer.gbuffer.AlbedoSpecularBuffer)

	gl.DrawArrays(gl.TRIANGLES, 0, 3)
}

func (renderer *Renderer) PointLightPass() {

}

func (renderer *Renderer) SpotLightPass() {

}

func (renderer *Renderer) ForwardPass() {
	renderer.gbuffer.BindReadOnly()
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0)
	gl.BlitFramebuffer(
		0,
		0,
		int32(renderer.width),
		int32(renderer.height),
		0,
		0,
		int32(renderer.width),
		int32(renderer.height),
		gl.DEPTH_BUFFER_BIT,
		gl.NEAREST)
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}