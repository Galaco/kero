package renderer

import (
	"github.com/galaco/gosigl"
	"github.com/galaco/kero/event/message"
	"github.com/galaco/kero/framework/graphics"
	graphics3d "github.com/galaco/kero/framework/graphics/3d"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/systems"
	"github.com/galaco/kero/systems/renderer/cache"
	"github.com/galaco/kero/systems/renderer/shaders"
)

type Renderer struct {
	systems.System

	materialCache *cache.Material
	textureCache *cache.Texture
	shaderCache *cache.Shader

	gpuItemCache *cache.GpuItem

	scene *SceneGraph
}

func (s *Renderer) Register() {
	s.shaderCache,_ = shaders.LoadShaders()

	gosigl.EnableBlend()
	gosigl.EnableDepthTest()
	gosigl.EnableCullFace(gosigl.Back, gosigl.WindingClockwise)
}

func (s *Renderer) Update(dt float64) {
	if s.scene == nil {
		return
	}
	s.startFrame()
	s.renderBsp()
}

func (s *Renderer) ProcessMessage(message message.Dispatchable) {
	switch message.Type() {
	case messages.TypeLoadingLevelParsed:
		s.scene = NewSceneGraphFromBsp(message.(*messages.LoadingLevelParsed).Level(), s.materialCache, s.textureCache, s.gpuItemCache)
	}
}

func (s *Renderer) startFrame() {
	activeShader := s.shaderCache.Find("LightMappedGeneric")

	projection := graphics3d.NewCamera(90, 16/9).ProjectionMatrix()
	view := graphics3d.NewCamera(90, 16/9).ViewMatrix()

	activeShader.Bind()
	graphics.PushMat4(activeShader.GetUniform("projection"), 1, false, projection)
	graphics.PushMat4(activeShader.GetUniform("view"), 1, false, view)
}

func (s *Renderer) renderBsp() {
	graphics.BindMesh(&s.scene.gpuMesh)
	for _,f := range s.scene.bspFaces {
		graphics.DrawFace(f.Offset(), f.Length(), s.gpuItemCache.Find(f.Material()))
	}
}

func NewRenderer() *Renderer {
	return &Renderer{
		textureCache: cache.NewTextureCache(),
		materialCache: cache.NewMaterialCache(),
		gpuItemCache: cache.NewGpuItemCache(),
	}
}