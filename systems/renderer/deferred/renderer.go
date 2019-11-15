package deferred

import (
	"errors"
	"github.com/galaco/gosigl"
	"github.com/galaco/kero/framework/graphics"
	graphics3d "github.com/galaco/kero/framework/graphics/3d"
	"github.com/galaco/kero/framework/graphics/adapter"
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
	adapter.PushInt32(renderer.geometryShader.GetUniform("normalSampler"), 1)

	renderer.directionalLightShader = graphics.NewShader()
	if err := renderer.directionalLightShader.Add(gosigl.VertexShader, DirectionalLightPassVertex); err != nil {
		return err
	}
	if err := renderer.directionalLightShader.Add(gosigl.FragmentShader, DirectionalLightPassFragment); err != nil {
		return err
	}
	renderer.directionalLightShader.Finish()

	gosigl.EnableDepthTest()
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
	adapter.ClearColor(0, 0, 0, 1)
	adapter.ClearAll()

	renderer.bindShader(renderer.geometryShader)

	adapter.PushMat4(renderer.geometryShader.GetUniform("projection"), 1, false, camera.ProjectionMatrix())
	adapter.PushMat4(renderer.geometryShader.GetUniform("view"), 1, false, camera.ViewMatrix())
	adapter.PushMat4(renderer.geometryShader.GetUniform("model"), 1, false, camera.ModelMatrix())
	adapter.PushInt32(renderer.geometryShader.GetUniform("albedoSampler"), 0)

	gosigl.EnableCullFace(gosigl.Back, gosigl.WindingClockwise)
}

func (renderer *Renderer) DirectionalLightPass(light *DirectionalLight) {
	gosigl.EnableCullFace(gosigl.Back, gosigl.WindingCounterClockwise)

	adapter.BindFrameBuffer(0)
	adapter.ClearAll()

	renderer.bindShader(renderer.directionalLightShader)

	adapter.PushInt32(renderer.directionalLightShader.GetUniform("uPositionTex"), 0)
	adapter.BindTextureToSlot(0, renderer.gbuffer.PositionBuffer)

	adapter.PushInt32(renderer.directionalLightShader.GetUniform("uNormalTex"), 1)
	adapter.BindTextureToSlot(1, renderer.gbuffer.NormalBuffer)

	adapter.PushInt32(renderer.directionalLightShader.GetUniform("uColorTex"), 2)
	adapter.BindTextureToSlot(2, renderer.gbuffer.AlbedoSpecularBuffer)

	adapter.PushVec3(
		renderer.directionalLightShader.GetUniform("directionalLight.Base.Color"),
		light.Color.X(), light.Color.Y(), light.Color.Z())
	adapter.PushFloat32(
		renderer.directionalLightShader.GetUniform("directionalLight.Base.DiffuseIntensity"),
		light.DiffuseIntensity)

	adapter.PushVec3(
		renderer.directionalLightShader.GetUniform("directionalLight.AmbientColor"),
		light.AmbientColor.X(), light.AmbientColor.Y(), light.AmbientColor.Z())
	adapter.PushFloat32(
		renderer.directionalLightShader.GetUniform("directionalLight.AmbientIntensity"),
		light.AmbientIntensity)
	normalizedDirection := light.Direction.Normalize()
	adapter.PushVec3(
		renderer.directionalLightShader.GetUniform("directionalLight.Direction"),
		normalizedDirection.X(), normalizedDirection.Y(), normalizedDirection.Z())

	adapter.DrawArray(0, 3)
}

func (renderer *Renderer) PointLightPass() {

}

func (renderer *Renderer) SpotLightPass() {

}

func (renderer *Renderer) ForwardPass() {
	renderer.gbuffer.BindReadOnly()
	adapter.BindFrameBufferDraw(0)
	adapter.BlitDepthBuffer(int32(renderer.width), int32(renderer.height))
	adapter.BindFrameBuffer(0)

	gosigl.EnableCullFace(gosigl.Back, gosigl.WindingClockwise)
}
