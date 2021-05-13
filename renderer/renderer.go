package renderer

import (
	"errors"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/entity"
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/framework/graphics/adapter"
	"github.com/galaco/kero/framework/graphics/mesh"
	"github.com/galaco/kero/framework/physics/raytrace"
	scene2 "github.com/galaco/kero/framework/scene"
	"github.com/galaco/kero/framework/scene/vis"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/renderer/cache"
	"github.com/galaco/kero/renderer/scene"
	"github.com/galaco/kero/renderer/shaders"
	"github.com/galaco/kero/utils"
	"github.com/go-gl/mathgl/mgl32"
	"math"
	"strings"
)

type Renderer struct {
	shaderCache *cache.Shader

	dataScene *scene2.StaticScene
	gpuScene  scene.GPUScene

	activeShader *adapter.Shader

	flags struct {
		matLeafvis int32
	}
}

func (s *Renderer) Initialize() {
	var err error
	s.shaderCache, err = shaders.LoadShaders()
	if err != nil {
		panic(err)
	}

	adapter.EnableBlending()
	adapter.EnableDepthTesting()
	adapter.EnableBackFaceCulling()

	event.Get().AddListener(messages.TypeLoadingLevelParsed, s.onLoadingLevelParsed)
	s.bindConVars()
}

func (s *Renderer) Render() {
	if s.dataScene == nil {
		return
	}
	s.dataScene.RecomputeVisibleClusters()
	clusters := s.computeRenderableClusters(graphics.FrustumFromCamera(s.dataScene.Camera))

	// Draw skybox
	// Skip sky rendering if all renderable clusters cannot see the sky or we are outside the map
	var shouldRenderSkybox bool
	if s.gpuScene.Skybox != nil && s.dataScene.CurrentLeaf != nil && s.dataScene.CurrentLeaf.Cluster != -1 {
		for _, c := range clusters {
			if c.SkyVisible {
				shouldRenderSkybox = true
				break
			}
		}
	}

	if shouldRenderSkybox {
		s.renderSkybox(s.gpuScene.Skybox)
		if s.dataScene.SkyCamera != nil {
			origin := s.dataScene.SkyCamera.Transform().Position
			s.dataScene.SkyCamera.Transform().Rotation = s.dataScene.Camera.Transform().Rotation
			s.dataScene.SkyCamera.Transform().Position = s.dataScene.SkyCamera.Transform().Position.Add(s.dataScene.Camera.Transform().Position.Mul(1 / s.dataScene.SkyCamera.Transform().Scale.X()))
			s.dataScene.SkyCamera.Update(0)
			s.startFrame(s.dataScene.SkyCamera)
			s.renderBsp(s.dataScene.SkyCamera, s.dataScene.SkyboxClusterLeafs)
			s.renderDisplacements(s.dataScene.DisplacementFaces)
			s.renderStaticProps(s.dataScene.SkyCamera, s.dataScene.SkyboxClusterLeafs)
			adapter.ClearDepthBuffer()
			s.dataScene.SkyCamera.Transform().Position = origin
		}
	}

	// Draw world
	s.startFrame(s.dataScene.Camera)
	s.renderBsp(s.dataScene.Camera, clusters)
	s.renderDisplacements(s.dataScene.DisplacementFaces)
	s.renderStaticProps(s.dataScene.Camera, clusters)
	s.renderEntityProps()

	s.DrawDebug()
}

func (s *Renderer) DrawDebug() {
	debugPoints := make([]float32, 0)
	switch console.GetConvarInt("mat_leafvis") {
	case 1:
		for _, l := range s.dataScene.ClusterLeafs {
			debugPoints = append(debugPoints, mesh.NewCuboidFromMinMaxs(mgl32.Vec3{l.Mins.X(), l.Mins.Y(), l.Mins.Z()}, mgl32.Vec3{l.Maxs.X(), l.Maxs.Y(), l.Maxs.Z()}).Vertices()...)
		}
	case 2:
		if s.dataScene.CurrentLeaf != nil {
			debugPoints = append(debugPoints, mesh.NewCuboidFromMinMaxs(
				mgl32.Vec3{
					float32(s.dataScene.CurrentLeaf.Mins[0]),
					float32(s.dataScene.CurrentLeaf.Mins[1]),
					float32(s.dataScene.CurrentLeaf.Mins[2]),
				},
				mgl32.Vec3{
					float32(s.dataScene.CurrentLeaf.Maxs[0]),
					float32(s.dataScene.CurrentLeaf.Maxs[1]),
					float32(s.dataScene.CurrentLeaf.Maxs[2]),
				},
			).Vertices()...)
		}
	case 3:
		for _, l := range s.dataScene.VisibleClusterLeafs {
			debugPoints = append(debugPoints, mesh.NewCuboidFromMinMaxs(mgl32.Vec3{l.Mins.X(), l.Mins.Y(), l.Mins.Z()}, mgl32.Vec3{l.Maxs.X(), l.Maxs.Y(), l.Maxs.Z()}).Vertices()...)
		}
	}
	adapter.PushMat4(s.activeShader.GetUniform("model"), 1, false, s.dataScene.Camera.ModelMatrix())

	var testEnt entity.IEntity
	for _, e := range s.dataScene.Entities {
		if strings.HasPrefix(e.Classname(), "prop_") {
			testEnt = e
			break
		}
	}
	if testEnt != nil {
		result := raytrace.TraceRayBetween(s.dataScene, s.dataScene.Camera.Transform().Position, testEnt.Transform().Position)

		if result.Hit {
			adapter.DrawLine(testEnt.Transform().Position, result.Point, mgl32.Vec3{0, 255, 0})
		}
		//adapter.DrawLine(s.dataScene.Camera.Transform().Position.Add(mgl32.Vec3{1,1,1}), testEnt.Transform().Position, mgl32.Vec3{0,255,0})
		//adapter.DrawLine(mgl32.Vec3{0,0,0}, testEnt.Transform().Position, mgl32.Vec3{0,255,0})
	}

	adapter.DrawDebugLines(debugPoints, mgl32.Vec3{0, 255, 0})
}

func (s *Renderer) FinishFrame() {
	adapter.ClearColor(0.25, 0.25, 0.25, 1)
	adapter.ClearAll()
}

func (s *Renderer) onLoadingLevelParsed(message interface{}) {
	s.dataScene = message.(*messages.LoadingLevelParsed).Level().(*scene2.StaticScene)
	s.gpuScene = *scene.GpuSceneFromFrameworkScene(s.dataScene, filesystem.Get())
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
	if console.GetConvarBoolean("r_drawlightmaps") == true {
		adapter.PushInt32(s.activeShader.GetUniform("renderLightmapsAsAlbedo"), 1)
	} else {
		adapter.PushInt32(s.activeShader.GetUniform("renderLightmapsAsAlbedo"), 0)
	}

	adapter.BindMesh(&s.gpuScene.GpuMesh)
	adapter.PushInt32(s.activeShader.GetUniform("albedoSampler"), 0)
	adapter.PushInt32(s.activeShader.GetUniform("lightmapSampler"), 4)
	adapter.BindLightmap(s.gpuScene.GpuItemCache.Find(scene2.LightmapTexturePath))
	var mat *cache.GpuMaterial

	materialMappedClusterFaces := vis.GroupClusterFacesByMaterial(clusters)

	// SORTING
	opaqueMaterials := map[*cache.GpuMaterial][]*graphics.BspFace{}
	translucentMaterials := map[*cache.GpuMaterial][]*graphics.BspFace{}

	for clusterFaceMaterial, faces := range materialMappedClusterFaces {
		mat = s.gpuScene.GpuMaterialCache.Find(clusterFaceMaterial)

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
		indices = append(indices, s.dataScene.BspMesh.Indices()[face.Offset():face.Offset()+(face.Length())]...)
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
		mat = s.gpuScene.GpuMaterialCache.Find(displacement.Material())
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
			if gpuProp, ok := s.gpuScene.GpuStaticProps[prop.Model().Id]; ok {
				for idx := range gpuProp.Id {
					adapter.BindMesh(&gpuProp.Id[idx])
					adapter.BindTexture(gpuProp.Material[idx].Diffuse)
					adapter.DrawIndexedArray(len(prop.Model().Meshes()[idx].Indices()), 0, nil)
				}
			}
		}
	}
}

func (s *Renderer) renderEntityProps() {
	for _, entry := range s.gpuScene.GpuRenderablePropEntities {
		for _, ent := range entry.Entities {
			adapter.PushMat4(s.activeShader.GetUniform("model"), 1, false, ent.Transform().TransformationMatrix())
			if gpuProp, ok := s.gpuScene.GpuStaticProps[entry.Id]; ok {
				for idx := range gpuProp.Id {
					adapter.BindMesh(&gpuProp.Id[idx])
					adapter.BindTexture(gpuProp.Material[idx].Diffuse)
					adapter.DrawIndexedArray(len(entry.Prop.Meshes()[idx].Indices()), 0, nil)
				}
			}
		}
	}
}

func (s *Renderer) computeRenderableClusters(viewFrustum *graphics.Frustum) []*vis.ClusterLeaf {
	renderClusters := make([]*vis.ClusterLeaf, 0, 64)
	for idx, cluster := range s.dataScene.VisibleClusterLeafs {
		if !viewFrustum.IsLeafInFrustum(cluster.Mins, cluster.Maxs) {
			continue
		}
		renderClusters = append(renderClusters, s.dataScene.VisibleClusterLeafs[idx])
	}
	return renderClusters
}

func (s *Renderer) renderSkybox(skybox *scene.Skybox) {
	skyboxTransform := skybox.SkyMeshTransform
	skyboxTransform.Position = s.dataScene.Camera.Transform().Position

	s.activeShader = s.shaderCache.Find("Skybox")
	s.activeShader.Bind()
	adapter.PushInt32(s.activeShader.GetUniform("albedoSampler"), 0)
	adapter.PushMat4(s.activeShader.GetUniform("projection"), 1, false, s.dataScene.Camera.ProjectionMatrix())
	adapter.PushMat4(s.activeShader.GetUniform("view"), 1, false, s.dataScene.Camera.ViewMatrix())
	adapter.PushMat4(s.activeShader.GetUniform("model"), 1, false, skyboxTransform.TransformationMatrix())

	adapter.BindMesh(&skybox.SkyMeshGpuID)
	adapter.BindCubemap(skybox.SkyMaterialGpuID)
	adapter.DrawArray(0, len(skybox.SkyMesh.Vertices()))
}

func (s *Renderer) ReleaseGPUResources() {

}

func (s *Renderer) bindConVars() {
	console.AddConvarBool("r_drawlightmaps", "Render lightmaps as diffuse material", false)
	console.AddConvarInt("mat_leafvis", "Render visleaf wireframes", 0)

	// Currently broken (texcache is flushed after gpu upload so raw lightmap colour data is unavailable)
	console.AddCommand("kero_dumplightmap", "Dump lightmap texture to a JPG", "kero_dumplightmap <filepath/filename>", func(options string) error {
		if s == nil {
			return nil
		}

		if ok := s.dataScene.TexCache.Find(scene2.LightmapTexturePath); ok != nil {
			utils.DumpLightmap(options, ok)
			return nil
		}

		return errors.New("kero_dumplightmap: no lightmap in memory")
	})
	console.AddCommand("kero_drawlightmaps", "Renders lightmaps in place of diffuse textures", "kero_drawlightmaps <0|1>", func(options string) error {
		if s == nil {
			return nil
		}
		if ok := s.dataScene.TexCache.Find(scene2.LightmapTexturePath); ok == nil {
			return errors.New("kero_drawlightmaps: no lightmap in memory")
		}
		if options == "1" {
			console.SetConvarBoolean("r_drawlightmaps", true)
		} else {
			console.SetConvarBoolean("r_drawlightmaps", false)
		}
		return nil
	})
}

func NewRenderer() *Renderer {
	return &Renderer{}
}
