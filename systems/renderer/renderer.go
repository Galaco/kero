package renderer

import (
	"github.com/galaco/gosigl"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/entity"
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/systems"
	"github.com/galaco/kero/systems/renderer/cache"
	"github.com/galaco/kero/systems/renderer/scene"
	"github.com/galaco/kero/systems/renderer/shaders"
	"github.com/galaco/kero/systems/renderer/vis"
	"github.com/galaco/kero/valve"
	"math"
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
	s.renderDisplacements(s.scene.displacementFaces)
	s.renderStaticProps(clusters)

	s.renderSkybox(clusters, s.scene.skybox)
}

func (s *Renderer) ProcessMessage(message event.Dispatchable) {
	switch message.Type() {
	case messages.TypeLoadingLevelParsed:
		s.scene = NewSceneGraphFromBsp(
			s.context.Filesystem,
			message.(*messages.LoadingLevelParsed).Level().(*valve.Bsp),
			message.(*messages.LoadingLevelParsed).Entities().([]entity.Entity),
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

func (s *Renderer) renderDisplacements(displacements []*valve.BspFace) {
	var mat *cache.GpuMaterial
	for _, displacement := range displacements {
		mat = s.materialCache.Find(displacement.Material())
		graphics.DrawFace(displacement.Offset(), displacement.Length(), mat.Diffuse)
		if err := graphics.GpuError(); err != nil {
			event.Dispatch(messages.NewConsoleMessage(console.LevelError, err.Error()))
		}
	}
}

func (s *Renderer) renderStaticProps(clusters []*vis.ClusterLeaf) {
	viewPosition := s.scene.camera.Transform().Position

	for _, cluster := range clusters {
		distToCluster := math.Pow(float64(cluster.Origin.X()-viewPosition.X()), 2) +
			math.Pow(float64(cluster.Origin.Y()-viewPosition.Y()), 2) +
			math.Pow(float64(cluster.Origin.Z()-viewPosition.Z()), 2)

		for _, prop := range cluster.StaticProps {
			//  Skip render if staticProp is fully faded
			if prop.FadeMaxDistance() > 0 && distToCluster >= math.Pow(float64(prop.FadeMaxDistance()), 2) {
				continue
			}
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

func (s *Renderer) renderSkybox(clusters []*vis.ClusterLeaf, skybox *scene.Skybox) {
	// Skip sky rendering if all renderable clusters cannot see the sky
	var isVisible bool
	for _,c := range clusters {
		if c.SkyVisible{
			isVisible = true
			break
		}
	}
	if !isVisible {
		return
	}

	skyboxTransform := skybox.SkyMeshTransform
	skyboxTransform.Position = s.scene.camera.Transform().Position

	s.activeShader = s.shaderCache.Find("Skybox")
	s.activeShader.Bind()
	graphics.PushInt32(s.activeShader.GetUniform("albedoSampler"), 0)
	graphics.PushMat4(s.activeShader.GetUniform("projection"), 1, false, s.scene.camera.ProjectionMatrix())
	graphics.PushMat4(s.activeShader.GetUniform("view"), 1, false, s.scene.camera.ViewMatrix())
	graphics.PushMat4(s.activeShader.GetUniform("model"), 1, false, skyboxTransform.TransformationMatrix())

	//gosigl.EnableDepthTest()
	//gosigl.EnableCullFace(gosigl.Front, gosigl.WindingClockwise)

	graphics.BindMesh(&skybox.SkyMeshGpuID)
	graphics.BindCubemap(skybox.SkyMaterialGpuID)
	graphics.DrawArray(0, len(skybox.SkyMesh.Vertices()))
	//
	//gosigl.EnableBlend()
	//gosigl.EnableDepthTest()
	//gosigl.EnableCullFace(gosigl.Back, gosigl.WindingClockwise)
}

func NewRenderer() *Renderer {
	return &Renderer{
		textureCache:   cache.NewTextureCache(),
		materialCache:  cache.NewMaterialCache(),
		gpuItemCache:   cache.NewGpuItemCache(),
		gpuStaticProps: map[string]*cache.GpuProp{},
	}
}
