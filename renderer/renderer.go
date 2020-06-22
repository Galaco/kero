package renderer

import (
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/entity"
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/framework/graphics"
	graphics3d "github.com/galaco/kero/framework/graphics/3d"
	"github.com/galaco/kero/library/valve"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/renderer/cache"
	"github.com/galaco/kero/renderer/scene"
	"github.com/galaco/kero/renderer/shaders"
	"github.com/galaco/kero/renderer/vis"
	"math"
)

type Renderer struct {
	materialCache  *cache.Material
	textureCache   *cache.Texture
	shaderCache    *cache.Shader
	gpuStaticProps map[string]*cache.GpuProp

	gpuItemCache *cache.GpuItem
	activeShader *graphics.Shader

	scene *StaticScene
}

func (s *Renderer) Initialize() {
	var err error
	s.shaderCache, err = shaders.LoadShaders()
	if err != nil {
		panic(err)
	}

	graphics.EnableBlending()
	graphics.EnableDepthTesting()
	graphics.EnableBackFaceCulling()

	event.Get().AddListener(messages.TypeLoadingLevelParsed, s.onLoadingLevelParsed)
}

func (s *Renderer) Render() {
	if s.scene == nil {
		return
	}
	s.scene.RecomputeVisibleClusters()
	clusters := s.computeRenderableClusters(vis.FrustumFromCamera(s.scene.camera))

	// Draw skybox

	s.renderSkybox(clusters, s.scene.skybox)

	// Draw world
	s.startFrame(s.scene.camera)
	s.renderBsp(s.scene.camera, clusters)
	s.renderDisplacements(s.scene.displacementFaces)
	s.renderStaticProps(s.scene.camera, clusters)
}

func (s *Renderer) onLoadingLevelParsed(message interface{}) {
	s.scene = NewStaticSceneFromBsp(
		filesystem.Get(),
		message.(*messages.LoadingLevelParsed).Level().(*valve.Bsp),
		message.(*messages.LoadingLevelParsed).Entities().([]entity.IEntity),
		s.materialCache,
		s.textureCache,
		s.gpuItemCache,
		s.gpuStaticProps)
}

func (s *Renderer) startFrame(camera *graphics3d.Camera) {
	projection := s.scene.camera.ProjectionMatrix()
	view := camera.ViewMatrix()

	s.activeShader = s.shaderCache.Find("LightMappedGeneric")
	s.activeShader.Bind()
	graphics.PushMat4(s.activeShader.GetUniform("projection"), 1, false, projection)
	graphics.PushMat4(s.activeShader.GetUniform("view"), 1, false, view)
}

func (s *Renderer) renderBsp(camera *graphics3d.Camera, clusters []*vis.ClusterLeaf) {
	graphics.PushMat4(s.activeShader.GetUniform("model"), 1, false, camera.ModelMatrix())

	graphics.BindMesh(&s.scene.gpuMesh)
	graphics.PushInt32(s.activeShader.GetUniform("albedoSampler"), 0)
	var mat *cache.GpuMaterial

	materialMappedClusterFaces := vis.GroupClusterFacesByMaterial(clusters)

	// SORTING
	opaqueMaterials := map[*cache.GpuMaterial][]*valve.BspFace{}
	translucentMaterials := map[*cache.GpuMaterial][]*valve.BspFace{}

	for clusterFaceMaterial, faces := range materialMappedClusterFaces {
		mat = s.materialCache.Find(clusterFaceMaterial)

		if mat.Properties.Skip {
			continue
		}

		if mat.Properties.Translucent || mat.Properties.Alpha > 0 {
			translucentMaterials[mat] = faces
		} else {
			opaqueMaterials[mat] = faces
		}
	}

	for clusterFaceMaterial, faces := range opaqueMaterials {
		s.RenderBSPMaterial(clusterFaceMaterial, faces)
	}

	graphics.PushInt32(s.activeShader.GetUniform("hasTranslucentProperty"), 1)

	for clusterFaceMaterial, faces := range translucentMaterials {
		graphics.PushFloat32(s.activeShader.GetUniform("alpha"), clusterFaceMaterial.Properties.Alpha)
		if clusterFaceMaterial.Properties.Translucent == true {
			graphics.PushInt32(s.activeShader.GetUniform("translucent"), 1)
		} else {
			graphics.PushInt32(s.activeShader.GetUniform("translucent"), 0)
		}
		s.RenderBSPMaterial(clusterFaceMaterial, faces)
	}
	graphics.PushInt32(s.activeShader.GetUniform("hasTranslucentProperty"), 0)
}

func (s *Renderer) RenderBSPMaterial(mat *cache.GpuMaterial, faces []*valve.BspFace) {
	indices := make([]uint32, 0, 256)
	for _, face := range faces {
		indices = append(indices, s.scene.bspMesh.Indices()[face.Offset():face.Offset()+(face.Length())]...)
	}
	graphics.UpdateIndexArrayBuffer(indices)
	graphics.BindTexture(mat.Diffuse)
	graphics.DrawIndexedArray(len(indices), 0, nil)
	if err := graphics.GpuError(); err != nil {
		console.PrintString(console.LevelError, err.Error())
	}
}

func (s *Renderer) renderDisplacements(displacements []*valve.BspFace) {
	var mat *cache.GpuMaterial
	for _, displacement := range displacements {
		mat = s.materialCache.Find(displacement.Material())
		graphics.DrawFace(displacement.Offset(), displacement.Length(), mat.Diffuse)
		if err := graphics.GpuError(); err != nil {
			console.PrintString(console.LevelError, err.Error())
		}
	}
}

func (s *Renderer) renderStaticProps(camera *graphics3d.Camera, clusters []*vis.ClusterLeaf) {
	viewPosition := camera.Transform().Position

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
					graphics.DrawIndexedArray(len(prop.Model().Meshes()[idx].Indices()), 0, nil)
				}
			}
		}
	}
}

func (s *Renderer) computeRenderableClusters(viewFrustum *vis.Frustum) []*vis.ClusterLeaf {
	renderClusters := make([]*vis.ClusterLeaf, 0, 64)
	for idx, cluster := range s.scene.visibleClusterLeafs {
		if !viewFrustum.IsLeafInFrustum(cluster.Mins, cluster.Maxs) {
			continue
		}
		renderClusters = append(renderClusters, s.scene.visibleClusterLeafs[idx])
	}
	return renderClusters
}

func (s *Renderer) renderSkybox(clusters []*vis.ClusterLeaf, skybox *scene.Skybox) {
	// Skip sky rendering if all renderable clusters cannot see the sky or we are outside the map
	if skybox == nil || s.scene.currentLeaf == nil || s.scene.currentLeaf.Cluster == -1 {
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
	skyboxTransform.Position = s.scene.camera.Transform().Position

	s.activeShader = s.shaderCache.Find("Skybox")
	s.activeShader.Bind()
	graphics.PushInt32(s.activeShader.GetUniform("albedoSampler"), 0)
	graphics.PushMat4(s.activeShader.GetUniform("projection"), 1, false, s.scene.camera.ProjectionMatrix())
	graphics.PushMat4(s.activeShader.GetUniform("view"), 1, false, s.scene.camera.ViewMatrix())
	graphics.PushMat4(s.activeShader.GetUniform("model"), 1, false, skyboxTransform.TransformationMatrix())

	//graphics.EnableDepthTesting()
	//graphics.EnableFrontFaceCulling()

	graphics.BindMesh(&skybox.SkyMeshGpuID)
	graphics.BindCubemap(skybox.SkyMaterialGpuID)
	graphics.DrawArray(0, len(skybox.SkyMesh.Vertices()))

	//graphics.EnableBlending()
	//graphics.EnableDepthTesting()
	//graphics.EnableBackFaceCulling()
}

func (s *Renderer) ReleaseGPUResources() {

}

func NewRenderer() *Renderer {
	return &Renderer{
		textureCache:   cache.NewTextureCache(),
		materialCache:  cache.NewMaterialCache(),
		gpuItemCache:   cache.NewGpuItemCache(),
		gpuStaticProps: map[string]*cache.GpuProp{},
	}
}
