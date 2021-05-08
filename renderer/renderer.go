package renderer

import (
	"errors"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/entity"
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/framework/graphics/adapter"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/renderer/cache"
	"github.com/galaco/kero/renderer/scene"
	"github.com/galaco/kero/renderer/shaders"
	"github.com/galaco/kero/renderer/vis"
	"github.com/galaco/kero/utils"
	"math"
)

type Renderer struct {
	materialCache  *cache.Material
	textureCache   *cache.Texture
	shaderCache    *cache.Shader
	gpuStaticProps map[string]*cache.GpuProp

	gpuItemCache *cache.GpuItem
	activeShader *adapter.Shader

	scene *StaticScene

	flags struct {
		renderLightmapsAsAlbedo   int32
		renderDebugLeafWireframes int32
	}
}

func (s *Renderer) Initialize() {
	var err error
	s.shaderCache, err = shaders.LoadShaders()
	if err != nil {
		panic(err)
	}

	s.textureCache.Add(cache.ErrorTexturePath, graphics.NewErrorTexture(cache.ErrorTexturePath))
	s.gpuItemCache.Add(cache.ErrorTexturePath, adapter.UploadTexture(s.textureCache.Find(cache.ErrorTexturePath)))

	adapter.EnableBlending()
	adapter.EnableDepthTesting()
	adapter.EnableBackFaceCulling()

	event.Get().AddListener(messages.TypeLoadingLevelParsed, s.onLoadingLevelParsed)
	s.bindConVars()
}

func (s *Renderer) Render() {
	if s.scene == nil {
		return
	}
	s.scene.RecomputeVisibleClusters()
	clusters := s.computeRenderableClusters(vis.FrustumFromCamera(s.scene.camera))

	// Draw skybox
	// Skip sky rendering if all renderable clusters cannot see the sky or we are outside the map
	var shouldRenderSkybox bool
	if s.scene.skybox != nil && s.scene.currentLeaf != nil && s.scene.currentLeaf.Cluster != -1 {
		for _, c := range clusters {
			if c.SkyVisible {
				shouldRenderSkybox = true
				break
			}
		}
	}

	if shouldRenderSkybox {
		s.renderSkybox(s.scene.skybox)
		if s.scene.skyCamera != nil {
			origin := s.scene.skyCamera.Transform().Position
			s.scene.skyCamera.Transform().Rotation = s.scene.camera.Transform().Rotation
			s.scene.skyCamera.Transform().Position = s.scene.skyCamera.Transform().Position.Add(s.scene.camera.Transform().Position.Mul(1 / s.scene.skyCamera.Transform().Scale.X()))
			s.scene.skyCamera.Update(0)
			s.startFrame(s.scene.skyCamera)
			s.renderBsp(s.scene.skyCamera, s.scene.skyboxClusterLeafs)
			s.renderDisplacements(s.scene.displacementFaces)
			s.renderStaticProps(s.scene.skyCamera, s.scene.skyboxClusterLeafs)
			adapter.ClearDepthBuffer()
			s.scene.skyCamera.Transform().Position = origin
		}
	}

	// Draw world
	s.startFrame(s.scene.camera)
	s.renderBsp(s.scene.camera, clusters)
	s.renderDisplacements(s.scene.displacementFaces)
	s.renderStaticProps(s.scene.camera, clusters)
}

func (s *Renderer) FinishFrame() {
	adapter.ClearColor(0.25, 0.25, 0.25, 1)
	adapter.ClearAll()
}

func (s *Renderer) onLoadingLevelParsed(message interface{}) {
	s.scene = NewStaticSceneFromBsp(
		filesystem.Get(),
		message.(*messages.LoadingLevelParsed).Level().(*graphics.Bsp),
		message.(*messages.LoadingLevelParsed).Entities().([]entity.IEntity),
		s.materialCache,
		s.textureCache,
		s.gpuItemCache,
		s.gpuStaticProps)
}

func (s *Renderer) startFrame(camera *graphics.Camera) {
	projection := camera.ProjectionMatrix()
	view := camera.ViewMatrix()

	s.activeShader = s.shaderCache.Find("LightMappedGeneric")
	s.activeShader.Bind()
	adapter.PushMat4(s.activeShader.GetUniform("projection"), 1, false, projection)
	adapter.PushMat4(s.activeShader.GetUniform("view"), 1, false, view)
}

func (s *Renderer) renderBsp(camera *graphics.Camera, clusters []*vis.ClusterLeaf) {
	adapter.PushMat4(s.activeShader.GetUniform("model"), 1, false, camera.ModelMatrix())
	adapter.PushInt32(s.activeShader.GetUniform("renderLightmapsAsAlbedo"), s.flags.renderLightmapsAsAlbedo)

	adapter.BindMesh(&s.scene.gpuMesh)
	adapter.PushInt32(s.activeShader.GetUniform("albedoSampler"), 0)
	adapter.PushInt32(s.activeShader.GetUniform("lightmapSampler"), 4)
	adapter.BindLightmap(s.gpuItemCache.Find(cache.LightmapTexturePath))
	var mat *cache.GpuMaterial

	materialMappedClusterFaces := vis.GroupClusterFacesByMaterial(clusters)

	// SORTING
	opaqueMaterials := map[*cache.GpuMaterial][]*graphics.BspFace{}
	translucentMaterials := map[*cache.GpuMaterial][]*graphics.BspFace{}

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

	adapter.PushInt32(s.activeShader.GetUniform("hasTranslucentProperty"), 1)

	for clusterFaceMaterial, faces := range translucentMaterials {
		adapter.PushFloat32(s.activeShader.GetUniform("alpha"), clusterFaceMaterial.Properties.Alpha)
		if clusterFaceMaterial.Properties.Translucent == true {
			adapter.PushInt32(s.activeShader.GetUniform("translucent"), 1)
		} else {
			adapter.PushInt32(s.activeShader.GetUniform("translucent"), 0)
		}
		s.RenderBSPMaterial(clusterFaceMaterial, faces)
	}
	adapter.PushInt32(s.activeShader.GetUniform("hasTranslucentProperty"), 0)
}

func (s *Renderer) RenderBSPMaterial(mat *cache.GpuMaterial, faces []*graphics.BspFace) {
	indices := make([]uint32, 0, 256)
	for _, face := range faces {
		indices = append(indices, s.scene.bspMesh.Indices()[face.Offset():face.Offset()+(face.Length())]...)
	}
	adapter.UpdateIndexArrayBuffer(indices)
	adapter.BindTexture(mat.Diffuse)
	adapter.DrawIndexedArray(len(indices), 0, nil)
	if err := adapter.GpuError(); err != nil {
		console.PrintString(console.LevelError, err.Error())
	}
}

func (s *Renderer) renderDisplacements(displacements []*graphics.BspFace) {
	var mat *cache.GpuMaterial
	for _, displacement := range displacements {
		mat = s.materialCache.Find(displacement.Material())
		adapter.DrawFace(displacement.Offset(), displacement.Length(), mat.Diffuse)
		if err := adapter.GpuError(); err != nil {
			console.PrintString(console.LevelError, err.Error())
		}
	}
}

func (s *Renderer) renderStaticProps(camera *graphics.Camera, clusters []*vis.ClusterLeaf) {
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
			adapter.PushMat4(s.activeShader.GetUniform("model"), 1, false, prop.Transform.TransformationMatrix())
			if gpuProp, ok := s.gpuStaticProps[prop.Model().Id]; ok {
				for idx := range gpuProp.Id {
					adapter.BindMesh(&gpuProp.Id[idx])
					adapter.BindTexture(gpuProp.Material[idx].Diffuse)
					adapter.DrawIndexedArray(len(prop.Model().Meshes()[idx].Indices()), 0, nil)
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

func (s *Renderer) renderSkybox(skybox *scene.Skybox) {
	skyboxTransform := skybox.SkyMeshTransform
	skyboxTransform.Position = s.scene.camera.Transform().Position

	s.activeShader = s.shaderCache.Find("Skybox")
	s.activeShader.Bind()
	adapter.PushInt32(s.activeShader.GetUniform("albedoSampler"), 0)
	adapter.PushMat4(s.activeShader.GetUniform("projection"), 1, false, s.scene.camera.ProjectionMatrix())
	adapter.PushMat4(s.activeShader.GetUniform("view"), 1, false, s.scene.camera.ViewMatrix())
	adapter.PushMat4(s.activeShader.GetUniform("model"), 1, false, skyboxTransform.TransformationMatrix())

	adapter.BindMesh(&skybox.SkyMeshGpuID)
	adapter.BindCubemap(skybox.SkyMaterialGpuID)
	adapter.DrawArray(0, len(skybox.SkyMesh.Vertices()))
}

func (s *Renderer) ReleaseGPUResources() {

}

func (s *Renderer) bindConVars() {
	console.AddCommand("kero_dumplightmap", "Dump lightmap texture to a JPG", "kero_dumplightmap <filepath/filename>", func(options string) error {
		if s == nil {
			return nil
		}

		if ok := s.textureCache.Find(cache.LightmapTexturePath); ok != nil {
			utils.DumpLightmap(options, ok)
			return nil
		}

		return errors.New("kero_dumplightmap: no lightmap in memory")
	})
	console.AddCommand("kero_drawlightmaps", "Renders lightmaps in place of diffuse textures", "kero_drawlightmaps <0|1>", func(options string) error {
		if s == nil {
			return nil
		}
		if ok := s.textureCache.Find(cache.LightmapTexturePath); ok == nil {
			return errors.New("kero_drawlightmaps: no lightmap in memory")
		}
		if options == "1" {
			s.flags.renderLightmapsAsAlbedo = 1
		} else {
			s.flags.renderLightmapsAsAlbedo = 0
		}
		return nil
	})
	console.AddCommand("kero_drawleafwireframes", "Renders visible leaf wireframes", "kero_drawleafwireframes <0|1>", func(options string) error {
		if s == nil {
			return nil
		}
		if options == "1" {
			s.flags.renderDebugLeafWireframes = 1
		} else {
			s.flags.renderDebugLeafWireframes = 0
		}
		return nil
	})
}

func NewRenderer() *Renderer {
	return &Renderer{
		textureCache:   cache.NewTextureCache(),
		materialCache:  cache.NewMaterialCache(),
		gpuItemCache:   cache.NewGpuItemCache(),
		gpuStaticProps: map[string]*cache.GpuProp{},
	}
}
