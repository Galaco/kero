package renderer

import (
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/entity"
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/graphics/adapter"
	"github.com/galaco/kero/framework/input"
	"github.com/galaco/kero/framework/window"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/systems"
	"github.com/galaco/kero/systems/renderer/cache"
	"github.com/galaco/kero/systems/renderer/deferred"
	"github.com/galaco/kero/systems/renderer/scene"
	"github.com/galaco/kero/systems/renderer/shaders"
	"github.com/galaco/kero/systems/renderer/vis"
	"github.com/galaco/kero/valve"
	"math"
)

type Renderer struct {
	context *systems.Context

	cache struct {
		materialCache *cache.Material
		textureCache  *cache.Texture
		shaderCache   *cache.Shader
	}

	gpu struct {
		staticProps map[string]*cache.GpuProp
		itemCache   *cache.GpuItem
	}

	deferred deferred.Renderer

	scene *SceneGraph
}

// Register
func (s *Renderer) Register(ctx *systems.Context) {
	win, err := window.CreateWindow(800, 600, "Kero")
	if err != nil {
		panic(err)
	}
	win.SetActive()
	input.SetBoundWindow(win)
	if err = adapter.Init(); err != nil {
		panic(err)
	}
	s.context = ctx
	s.cache.shaderCache, err = shaders.LoadShaders()
	if err != nil {
		panic(err)
	}

	if err = s.deferred.Init(win.Width(), win.Height()); err != nil {
		panic(err)
	}
}

// ProcessMessage
func (s *Renderer) ProcessMessage(message event.Dispatchable) {
	switch message.Type() {
	case messages.TypeLoadingLevelParsed:
		s.scene = NewSceneGraphFromBsp(
			s.context.Filesystem,
			message.(*messages.LoadingLevelParsed).Level().(*valve.Bsp),
			message.(*messages.LoadingLevelParsed).Entities().([]entity.Entity),
			s.cache.materialCache,
			s.cache.textureCache,
			s.gpu.itemCache,
			s.gpu.staticProps)
	}
}

// Update
func (s *Renderer) Update(dt float64) {
	if s.scene == nil {
		return
	}
	s.scene.RecomputeVisibleClusters(s.context.Client.Camera())
	s.DrawFrame(s.computeRenderableClusters(vis.FrustumFromCamera(s.context.Client.Camera())))
}

// DrawFrame
func (s *Renderer) DrawFrame(visibleClusters []*vis.ClusterLeaf) {

	s.deferred.GeometryPass(s.context.Client.Camera())
	s.renderBsp(visibleClusters)
	s.renderDisplacements(s.scene.displacementFaces)
	s.renderStaticProps(visibleClusters)

	s.deferred.DirectionalLightPass(s.scene.lightEnvironment)

	s.deferred.PointLightPass()
	// render point lights

	s.deferred.SpotLightPass()
	// render spot lights

	s.deferred.ForwardPass()
	//s.renderSkybox(visibleClusters, s.scene.skybox)
}

// renderBsp
func (s *Renderer) renderBsp(clusters []*vis.ClusterLeaf) {
	adapter.BindMesh(&s.scene.gpuMesh)
	var mat *cache.GpuMaterial

	materialMappedClusterFaces := vis.GroupClusterFacesByMaterial(clusters)
	var hasNormalMap int32
	for clusterFaceMaterial, faces := range materialMappedClusterFaces {
		mat = s.cache.materialCache.Find(clusterFaceMaterial)

		if mat.Properties.Skip {
			continue
		}

		adapter.BindTexture(mat.Diffuse)
		hasNormalMap = 0
		if mat.Properties.HasBumpMap {
			hasNormalMap = 1
		}
		adapter.PushInt32(s.deferred.ActiveShader().GetUniform("hasNormalSampler"), hasNormalMap)
		if mat.Properties.HasBumpMap {
			adapter.BindTextureToSlot(1, mat.Normal)
		}
		for _, face := range faces {
			adapter.DrawFace(face.Offset(), face.Length())
			if err := adapter.GpuError(); err != nil {
				event.Dispatch(messages.NewConsoleMessage(console.LevelError, err.Error()))
			}
		}
	}
	adapter.PushInt32(s.deferred.ActiveShader().GetUniform("hasNormalSampler"), 0)
}

// renderDisplacements
func (s *Renderer) renderDisplacements(displacements []*valve.BspFace) {
	var mat *cache.GpuMaterial
	var hasNormalMap int32
	for _, displacement := range displacements {
		mat = s.cache.materialCache.Find(displacement.Material())
		adapter.BindTexture(mat.Diffuse)
		hasNormalMap = 0
		if mat.Properties.HasBumpMap {
			hasNormalMap = 1
		}
		adapter.PushInt32(s.deferred.ActiveShader().GetUniform("hasNormalSampler"), hasNormalMap)
		if mat.Properties.HasBumpMap {
			adapter.BindTextureToSlot(1, mat.Normal)
		}
		adapter.DrawFace(displacement.Offset(), displacement.Length())
		if err := adapter.GpuError(); err != nil {
			event.Dispatch(messages.NewConsoleMessage(console.LevelError, err.Error()))
		}
	}
}

// renderStaticProps
func (s *Renderer) renderStaticProps(clusters []*vis.ClusterLeaf) {
	viewPosition := s.context.Client.Camera().Transform().Position
	var hasNormalMap int32

	for _, cluster := range clusters {
		distToCluster := math.Pow(float64(cluster.Origin.X()-viewPosition.X()), 2) +
			math.Pow(float64(cluster.Origin.Y()-viewPosition.Y()), 2) +
			math.Pow(float64(cluster.Origin.Z()-viewPosition.Z()), 2)

		for _, prop := range cluster.StaticProps {
			//  Skip render if staticProp is fully faded
			if prop.FadeMaxDistance() > 0 && distToCluster >= math.Pow(float64(prop.FadeMaxDistance()), 2) {
				continue
			}
			adapter.PushMat4(s.deferred.ActiveShader().GetUniform("model"), 1, false, prop.Transform.TransformationMatrix())
			if gpuProp, ok := s.gpu.staticProps[prop.Model().Id]; ok {
				for idx := range gpuProp.Id {
					adapter.BindMesh(gpuProp.Id[idx])
					adapter.BindTexture(gpuProp.Material[idx].Diffuse)
					hasNormalMap = 0
					if gpuProp.Material[idx].Properties.HasBumpMap {
						hasNormalMap = 1
					}
					adapter.PushInt32(s.deferred.ActiveShader().GetUniform("hasNormalSampler"), hasNormalMap)
					if gpuProp.Material[idx].Properties.HasBumpMap {
						adapter.BindTextureToSlot(1, gpuProp.Material[idx].Normal)
					}
					adapter.DrawArray(0, len(prop.Model().Meshes()[idx].Vertices()))
				}
			}
		}
	}
}

// computeRenderableClusters
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

// renderSkybox
func (s *Renderer) renderSkybox(clusters []*vis.ClusterLeaf, skybox *scene.Skybox) {
	// Skip sky rendering if all renderable clusters cannot see the sky or we are outside the map
	if s.scene.currentLeaf == nil || s.scene.currentLeaf.Cluster == -1 {
		return
	}
	var isVisible bool
	for _, c := range clusters {
		if c.SkyVisible {
			isVisible = true
			break
		}
	}
	if !isVisible {
		return
	}

	skyboxTransform := skybox.SkyMeshTransform
	skyboxTransform.Position = s.context.Client.Camera().Transform().Position

	shader := s.cache.shaderCache.Find("Skybox")
	shader.Bind()
	adapter.PushInt32(shader.GetUniform("albedoSampler"), 0)
	adapter.PushMat4(shader.GetUniform("projection"), 1, false, s.context.Client.Camera().ProjectionMatrix())
	adapter.PushMat4(shader.GetUniform("view"), 1, false, s.context.Client.Camera().ViewMatrix())
	adapter.PushMat4(shader.GetUniform("model"), 1, false, skyboxTransform.TransformationMatrix())

	adapter.BindMesh(&skybox.SkyMeshGpuID)
	adapter.BindCubemap(skybox.SkyMaterialGpuID)
	adapter.DrawArray(0, len(skybox.SkyMesh.Vertices()))
}

// NewRenderer
func NewRenderer() *Renderer {
	return &Renderer{
		cache: struct {
			materialCache *cache.Material
			textureCache  *cache.Texture
			shaderCache   *cache.Shader
		}{
			textureCache:  cache.NewTextureCache(),
			materialCache: cache.NewMaterialCache(),
		},
		gpu: struct {
			staticProps map[string]*cache.GpuProp
			itemCache   *cache.GpuItem
		}{
			itemCache:   cache.NewGpuItemCache(),
			staticProps: map[string]*cache.GpuProp{},
		},
	}
}
