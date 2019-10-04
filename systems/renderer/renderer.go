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
	context       *systems.Context
	materialCache *cache.Material
	textureCache  *cache.Texture
	shaderCache   *cache.Shader

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
	viewFrustum := vis.FrustumFromCamera(s.scene.camera)

	s.startFrame()
	s.renderBsp(viewFrustum)
}

func (s *Renderer) ProcessMessage(message event.Dispatchable) {
	switch message.Type() {
	case messages.TypeLoadingLevelParsed:
		s.scene = NewSceneGraphFromBsp(
			s.context.Filesystem,
			message.(*messages.LoadingLevelParsed).Level(),
			s.materialCache,
			s.textureCache,
			s.gpuItemCache)
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

func (s *Renderer) renderBsp(viewFrustum *vis.Frustum) {
	graphics.PushMat4(s.activeShader.GetUniform("model"), 1, false, s.scene.camera.ModelMatrix())

	graphics.BindMesh(&s.scene.gpuMesh)
	graphics.PushInt32(s.activeShader.GetUniform("albedoSampler"), 0)
	var mat *cache.GpuMaterial

	renderClusters := make([]*vis.ClusterLeaf, 0)
	for idx, cluster := range s.scene.visibleClusterLeafs {
		if !viewFrustum.IsCuboidInFrustum(cluster.Mins, cluster.Maxs) {
			continue
		}

		renderClusters = append(renderClusters, s.scene.visibleClusterLeafs[idx])
	}

	materialMappedClusterFaces := s.groupClusterFacesByMaterial(renderClusters)
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

// groupClusterFacesByMaterial groups all faces in a collections of
// clusters by material
func (s *Renderer) groupClusterFacesByMaterial(clusters []*vis.ClusterLeaf) map[string][]*valve.BspFace {
	clusterFaceMap := map[string][]*valve.BspFace{}

	for _, cluster := range clusters {
		for idx, face := range cluster.Faces {
			if _, ok := clusterFaceMap[face.Material()]; !ok {
				clusterFaceMap[face.Material()] = []*valve.BspFace{&cluster.Faces[idx]}
			} else {
				clusterFaceMap[face.Material()] = append(clusterFaceMap[face.Material()], &cluster.Faces[idx])
			}
		}
	}

	return clusterFaceMap
}

func NewRenderer() *Renderer {
	return &Renderer{
		textureCache:  cache.NewTextureCache(),
		materialCache: cache.NewMaterialCache(),
		gpuItemCache:  cache.NewGpuItemCache(),
	}
}
