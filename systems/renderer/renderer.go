package renderer

import (
	"github.com/galaco/gosigl"
	"github.com/galaco/kero/event"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/systems"
	"github.com/galaco/kero/systems/renderer/cache"
	"github.com/galaco/kero/systems/renderer/shaders"
	"github.com/galaco/kero/systems/renderer/vis"
	"github.com/galaco/kero/valve"
)

type Renderer struct {
	context        *systems.Context
	materialCache  *cache.Material
	textureCache   *cache.Texture
	shaderCache    *cache.Shader
	gpuStaticProps map[string]*cache.GpuProp

	gpuItemCache *cache.GpuItem
	activeShader *graphics.Shader

	scene *SceneGraph
}

func (s *Renderer) Register(ctx *systems.Context) {
	s.context = ctx
	var err error
	s.shaderCache, err = shaders.LoadShaders()
	if err != nil {
		panic(err)
	}

	gosigl.EnableBlend()
	gosigl.EnableDepthTest()
	gosigl.EnableCullFace(gosigl.Back, gosigl.WindingClockwise)
}

func (s *Renderer) Update(dt float64) {
	if s.scene == nil {
		return
	}
	s.scene.RecomputeVisibleClusters()
	clusters := s.computeRenderableClusters(vis.FrustumFromCamera(s.scene.camera))
	s.startFrame()
	s.renderBsp(clusters)
	s.renderStaticProps(clusters)
}

func (s *Renderer) ProcessMessage(message event.Dispatchable) {
	switch message.Type() {
	case messages.TypeLoadingLevelParsed:
		s.scene = NewSceneGraphFromBsp(
			s.context.Filesystem,
			message.(*messages.LoadingLevelParsed).Level().(*valve.Bsp),
			s.materialCache,
			s.textureCache,
			s.gpuItemCache,
			s.gpuStaticProps)
	}
}

func (s *Renderer) startFrame() {
	projection := s.scene.camera.ProjectionMatrix()
	view := s.scene.camera.ViewMatrix()

	s.activeShader = s.shaderCache.Find("LightMappedGeneric")
	s.activeShader.Bind()
	graphics.PushMat4(s.activeShader.GetUniform("projection"), 1, false, projection)
	graphics.PushMat4(s.activeShader.GetUniform("view"), 1, false, view)
}

func (s *Renderer) renderBsp(clusters []*vis.ClusterLeaf) {
	graphics.PushMat4(s.activeShader.GetUniform("model"), 1, false, s.scene.camera.ModelMatrix())

	graphics.BindMesh(&s.scene.gpuMesh)
	graphics.PushInt32(s.activeShader.GetUniform("albedoSampler"), 0)
	var mat *cache.GpuMaterial

	materialMappedClusterFaces := vis.GroupClusterFacesByMaterial(clusters)
	for clusterFaceMaterial, faces := range materialMappedClusterFaces {
		mat = s.materialCache.Find(clusterFaceMaterial)

		for _, face := range faces {
			graphics.DrawFace(face.Offset(), face.Length(), mat.Diffuse)
			if err := graphics.GpuError(); err != nil {
				event.Dispatch(messages.NewConsoleMessage(console.LevelError, err.Error()))
			}
		}
	}
}

func (s *Renderer) renderStaticProps(clusters []*vis.ClusterLeaf) {
	for _, cluster := range clusters {
		for _, prop := range cluster.StaticProps {
			graphics.PushMat4(s.activeShader.GetUniform("model"), 1, false, prop.Transform.TransformationMatrix())
			if gpuProp, ok := s.gpuStaticProps[prop.Model().Id]; ok {
				for idx := range gpuProp.Id {
					graphics.BindMesh(gpuProp.Id[idx])
					graphics.BindTexture(gpuProp.Material[idx].Diffuse)
					graphics.DrawArray(0, len(prop.Model().Meshes()[idx].Vertices()))
				}
			}
		}
	}
}

func (s *Renderer) computeRenderableClusters(viewFrustum *vis.Frustum) []*vis.ClusterLeaf {
	renderClusters := make([]*vis.ClusterLeaf, 0)
	for idx, cluster := range s.scene.visibleClusterLeafs {
		if !viewFrustum.IsCuboidInFrustum(cluster.Mins, cluster.Maxs) {
			continue
		}
		renderClusters = append(renderClusters, s.scene.visibleClusterLeafs[idx])
	}
	return renderClusters
}

func NewRenderer() *Renderer {
	return &Renderer{
		textureCache:   cache.NewTextureCache(),
		materialCache:  cache.NewMaterialCache(),
		gpuItemCache:   cache.NewGpuItemCache(),
		gpuStaticProps: map[string]*cache.GpuProp{},
	}
}
