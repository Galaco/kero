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
	pointLightShader	   *graphics.Shader
	activeShader           *graphics.Shader

	lightMesh 			   graphics.Mesh
	gpuLightMesh 		   adapter.GpuMesh
	lightMeshIndexVbo      uint32
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

	renderer.pointLightShader = graphics.NewShader()
	if err := renderer.pointLightShader.Add(gosigl.VertexShader,PointLightPassVertex); err != nil {
		return err
	}
	if err := renderer.pointLightShader.Add(gosigl.FragmentShader, PointLightPassFragment); err != nil {
		return err
	}
	renderer.pointLightShader.Finish()

	gosigl.EnableDepthTest()

	// Create light mesh
	renderer.lightMesh = graphics.NewSphere()
	renderer.gpuLightMesh, renderer.lightMeshIndexVbo = adapter.UploadLightMesh(renderer.lightMesh)

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

	renderer.bindPass(renderer.directionalLightShader)

	adapter.PushVec3(
		renderer.directionalLightShader.GetUniform("directionalLight.Base.Color"),
		light.Color.X(), light.Color.Y(), light.Color.Z())
	adapter.PushFloat32(
		renderer.directionalLightShader.GetUniform("directionalLight.Base.DiffuseIntensity"),
		light.DiffuseIntensity / 10)

	adapter.PushVec3(
		renderer.directionalLightShader.GetUniform("directionalLight.AmbientColor"),
		light.AmbientColor.X(), light.AmbientColor.Y(), light.AmbientColor.Z())
	adapter.PushFloat32(
		renderer.directionalLightShader.GetUniform("directionalLight.AmbientIntensity"),
		light.AmbientIntensity / 2)
	normalizedDirection := light.Direction.Normalize()
	adapter.PushVec3(
		renderer.directionalLightShader.GetUniform("directionalLight.Direction"),
		normalizedDirection.X(), normalizedDirection.Y(), normalizedDirection.Z())

	adapter.DrawArray(0, 3)
}

// PointLightPass
func (renderer *Renderer) PointLightPass(camera *graphics3d.Camera) {
	// @TODO TEST THESE PARAMS
	//gl.Disable(gl.DEPTH_TEST)
	//gl.Enable(gl.BLEND)
	//gl.BlendFunc(gl.ONE, gl.ONE)

	//gl.FrontFace(gl.CW)

	renderer.pointLightShader.Bind()

	renderer.bindShader(renderer.pointLightShader)
	renderer.bindPass(renderer.pointLightShader)
	adapter.PushVec3(renderer.pointLightShader.GetUniform( "uCameraPos"), camera.Transform().Position.X(), camera.Transform().Position.Y(), camera.Transform().Position.Z())

	adapter.PushMat4(renderer.pointLightShader.GetUniform("uVp"), 1, false, camera.ViewMatrix().Mul4(camera.ProjectionMatrix()))
	// We render every point light as a light sphere. And this light sphere is added onto the framebuffer
	// with additive alpha blending.
	adapter.BindMesh(&renderer.gpuLightMesh)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, renderer.lightMeshIndexVbo)

}

func (renderer *Renderer) SpotLightPass() {

}

func (renderer *Renderer) ForwardPass() {
	renderer.gbuffer.BindReadOnly()
	adapter.BindFrameBufferDraw(0)
	adapter.BlitDepthBuffer(renderer.width, renderer.height)
	adapter.BindFrameBuffer(0)

	gosigl.EnableCullFace(gosigl.Back, gosigl.WindingClockwise)
}

func (renderer *Renderer) RenderPointLight(light *PointLight) {
	adapter.PushFloat32(renderer.pointLightShader.GetUniform("uLightRadius"), light.DiffuseIntensity)
	adapter.PushVec3(renderer.pointLightShader.GetUniform("uLightPosition"), light.Position.X(), light.Position.Y(), light.Position.Z())
	adapter.PushVec3(renderer.pointLightShader.GetUniform("uLightColor"), light.Color.X(), light.Color.Y(), light.Color.Z())
	adapter.DrawElements(len(renderer.lightMesh.Indices()))
}

func (renderer *Renderer) bindPass(shader *graphics.Shader) {
	adapter.PushInt32(shader.GetUniform("uPositionTex"), 0)
	adapter.BindTextureToSlot(0, renderer.gbuffer.PositionBuffer)

	adapter.PushInt32(shader.GetUniform("uNormalTex"), 1)
	adapter.BindTextureToSlot(1, renderer.gbuffer.NormalBuffer)

	adapter.PushInt32(shader.GetUniform("uColorTex"), 2)
	adapter.BindTextureToSlot(2, renderer.gbuffer.AlbedoSpecularBuffer)
}