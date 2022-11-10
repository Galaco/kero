package renderer

import (
	"errors"
	"math"

	"github.com/galaco/kero/client/renderer/cache"
	"github.com/galaco/kero/client/renderer/scene"
	"github.com/galaco/kero/client/renderer/shaders"
	"github.com/galaco/kero/client/renderer/utils"
	"github.com/galaco/kero/internal/framework/console"
	"github.com/galaco/kero/internal/framework/event"
	"github.com/galaco/kero/internal/framework/filesystem"
	"github.com/galaco/kero/internal/framework/graphics"
	"github.com/galaco/kero/internal/framework/graphics/adapter"
	"github.com/galaco/kero/internal/framework/graphics/mesh"
	scene2 "github.com/galaco/kero/internal/framework/scene"
	"github.com/galaco/kero/internal/framework/scene/vis"
	"github.com/galaco/kero/shared/messages"
	scene3 "github.com/galaco/kero/shared/scene"
	"github.com/go-gl/mathgl/mgl32"
)

type Renderer struct {
	shaderCache *cache.Shader

	gpuScene scene.GPUScene

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
}

func (s *Renderer) Render() {
	if scene3.CurrentScene().Raw() == nil {
		return
	}
	scene3.CurrentScene().Raw().RecomputeVisibleClusters()
	clusters := s.computeRenderableClusters(graphics.FrustumFromCamera(scene3.CurrentScene().Raw().Camera))

	// Draw skybox
	// Skip sky rendering if all renderable clusters cannot see the sky or we are outside the map
	var shouldRenderSkybox bool
	if s.gpuScene.Skybox != nil && scene3.CurrentScene().Raw().CurrentLeaf != nil && scene3.CurrentScene().Raw().CurrentLeaf.Cluster != -1 {
		for _, c := range clusters {
			if c.SkyVisible {
				shouldRenderSkybox = true
				break
			}
		}
	}

	if shouldRenderSkybox {
		s.renderSkybox(s.gpuScene.Skybox)
		if scene3.CurrentScene().Raw().SkyCamera != nil {
			origin := scene3.CurrentScene().Raw().SkyCamera.Transform().Translation
			scene3.CurrentScene().Raw().SkyCamera.Transform().Orientation = scene3.CurrentScene().Raw().Camera.Transform().Orientation
			scene3.CurrentScene().Raw().SkyCamera.Transform().Translation = scene3.CurrentScene().Raw().SkyCamera.Transform().Translation.Add(scene3.CurrentScene().Raw().Camera.Transform().Translation.Mul(1 / scene3.CurrentScene().Raw().SkyCamera.Transform().Scale.X()))
			scene3.CurrentScene().Raw().SkyCamera.Update(0)
			s.startFrame(scene3.CurrentScene().Raw().SkyCamera)
			s.renderBsp(scene3.CurrentScene().Raw().SkyCamera, scene3.CurrentScene().Raw().SkyboxClusterLeafs)
			s.renderDisplacements(scene3.CurrentScene().Raw().DisplacementFaces)
			s.renderStaticProps(scene3.CurrentScene().Raw().SkyCamera, scene3.CurrentScene().Raw().SkyboxClusterLeafs)
			adapter.ClearDepthBuffer()
			scene3.CurrentScene().Raw().SkyCamera.Transform().Translation = origin
		}
	}

	// Draw world
	s.startFrame(scene3.CurrentScene().Raw().Camera)
	s.renderBsp(scene3.CurrentScene().Raw().Camera, clusters)
	s.renderDisplacements(scene3.CurrentScene().Raw().DisplacementFaces)
	s.renderStaticProps(scene3.CurrentScene().Raw().Camera, clusters)
	s.renderEntityProps()

	s.DrawDebug()
}

func (s *Renderer) DrawDebug() {
	debugPoints := make([]float32, 0)
	switch console.GetConvarInt("mat_leafvis") {
	case 1:
		for _, l := range scene3.CurrentScene().Raw().ClusterLeafs {
			debugPoints = append(debugPoints, mesh.NewCuboidFromMinMaxs(mgl32.Vec3{l.Mins.X(), l.Mins.Y(), l.Mins.Z()}, mgl32.Vec3{l.Maxs.X(), l.Maxs.Y(), l.Maxs.Z()}).Vertices()...)
		}
	case 2:
		if scene3.CurrentScene().Raw().CurrentLeaf != nil {
			debugPoints = append(debugPoints, mesh.NewCuboidFromMinMaxs(
				mgl32.Vec3{
					float32(scene3.CurrentScene().Raw().CurrentLeaf.Mins[0]),
					float32(scene3.CurrentScene().Raw().CurrentLeaf.Mins[1]),
					float32(scene3.CurrentScene().Raw().CurrentLeaf.Mins[2]),
				},
				mgl32.Vec3{
					float32(scene3.CurrentScene().Raw().CurrentLeaf.Maxs[0]),
					float32(scene3.CurrentScene().Raw().CurrentLeaf.Maxs[1]),
					float32(scene3.CurrentScene().Raw().CurrentLeaf.Maxs[2]),
				},
			).Vertices()...)
		}
	case 3:
		for _, l := range scene3.CurrentScene().Raw().VisibleClusterLeafs {
			debugPoints = append(debugPoints, mesh.NewCuboidFromMinMaxs(mgl32.Vec3{l.Mins.X(), l.Mins.Y(), l.Mins.Z()}, mgl32.Vec3{l.Maxs.X(), l.Maxs.Y(), l.Maxs.Z()}).Vertices()...)
		}
	}
	adapter.PushMat4(s.activeShader.GetUniform("model"), 1, false, scene3.CurrentScene().Raw().Camera.ModelMatrix())
	adapter.DrawDebugLines(debugPoints, mgl32.Vec3{0, 255, 0})

	if console.GetConvarBoolean("r_drawcollisionmodels") == true {
		s.drawCollisionMeshes()
	}
}

func (s *Renderer) FinishFrame() {
	adapter.ClearColor(0.25, 0.25, 0.25, 1)
	adapter.ClearAll()
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
		indices = append(indices, scene3.CurrentScene().Raw().BspMesh.Indices()[face.Offset():face.Offset()+(face.Length())]...)
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
	viewPosition := camera.Transform().Translation

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
			if gpuProp, ok := s.gpuScene.GpuStaticProps[prop.Model().Model.Id]; ok {
				for idx := range gpuProp.Id {
					adapter.BindMesh(&gpuProp.Id[idx])
					adapter.BindTexture(gpuProp.Material[idx].Diffuse)
					adapter.DrawIndexedArray(len(prop.Model().Model.Meshes()[idx].Indices()), 0, nil)
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
	for idx, cluster := range scene3.CurrentScene().Raw().VisibleClusterLeafs {
		if !viewFrustum.IsLeafInFrustum(cluster.Mins, cluster.Maxs) {
			continue
		}
		renderClusters = append(renderClusters, scene3.CurrentScene().Raw().VisibleClusterLeafs[idx])
	}
	return renderClusters
}

func (s *Renderer) renderSkybox(skybox *scene.Skybox) {
	skyboxTransform := skybox.SkyMeshTransform
	skyboxTransform.Translation = scene3.CurrentScene().Raw().Camera.Transform().Translation

	s.activeShader = s.shaderCache.Find("Skybox")
	s.activeShader.Bind()
	adapter.PushInt32(s.activeShader.GetUniform("albedoSampler"), 0)
	adapter.PushMat4(s.activeShader.GetUniform("projection"), 1, false, scene3.CurrentScene().Raw().Camera.ProjectionMatrix())
	adapter.PushMat4(s.activeShader.GetUniform("view"), 1, false, scene3.CurrentScene().Raw().Camera.ViewMatrix())
	adapter.PushMat4(s.activeShader.GetUniform("model"), 1, false, skyboxTransform.TransformationMatrix())

	adapter.BindMesh(&skybox.SkyMeshGpuID)
	adapter.BindCubemap(skybox.SkyMaterialGpuID)
	adapter.DrawArray(0, len(skybox.SkyMesh.Vertices()))
}

func (s *Renderer) Cleanup() {
	// Release GPU resources
	for _, s := range s.gpuScene.GpuStaticProps {
		for _, id := range s.Id {
			adapter.DeleteMeshResource(id)
		}
	}

	for _, id := range s.gpuScene.GpuItemCache.All() {
		adapter.DeleteTextureResource(id)
	}

	s.gpuScene = scene.GPUScene{}
}

func (s *Renderer) drawCollisionMeshes() {
	//	if adapter.CurrentShader() == nil {
	//		return
	//	}
	//	adapter.EnableFrontFaceCulling()
	//	adapter.DisableDepthTesting()
	//
	//	adapter.PushMat4(adapter.CurrentShader().GetUniform("model"), 1, false, mgl32.Ident4())
	//	verts := make([]float32, 0)
	//	for _, vert := range s.bspRigidBody.vertices {
	//		verts = append(verts, vert[0], vert[1], vert[2])
	//	}
	//	adapter.DrawDebugLines(verts, mgl32.Vec3{255, 0, 255})
	//
	//	for _, n := range s.physicsEntities {
	//		if n.Model().RigidBody == nil {
	//			continue
	//		}
	//		adapter.PushMat4(adapter.CurrentShader().GetUniform("model"), 1, false, n.Transform().TransformationMatrix())
	//		for _, r := range s.studiomodelCollisionMeshes[n.Model().Model.Id].vertices {
	//			verts := make([]float32, 0)
	//			for _, v := range r {
	//				verts = append(verts, v[0], v[1], v[2])
	//			}
	//			adapter.DrawDebugLines(verts, mgl32.Vec3{255, 0, 255})
	//		}
	//	}
	//	adapter.EnableDepthTesting()
	//	adapter.EnableBackFaceCulling()
}

func (s *Renderer) bindConVars() {
	console.AddConvarBool("r_drawlightmaps", "Render lightmaps as diffuse material", false)
	console.AddConvarBool("r_drawcollisionmodels", "Render collision mode vertices", false)
	console.AddConvarInt("mat_leafvis", "Render visleaf wireframes", 0)

	// Currently broken (texcache is flushed after gpu upload so raw lightmap colour data is unavailable)
	console.AddCommand("kero_dumplightmap", "Dump lightmap texture to a JPG", "kero_dumplightmap <filepath/filename>", func(options string) error {
		if s == nil {
			return nil
		}

		if ok := scene3.CurrentScene().Raw().TexCache.Find(scene2.LightmapTexturePath); ok != nil {
			utils.DumpLightmap(options, ok)
			return nil
		}

		return errors.New("kero_dumplightmap: no lightmap in memory")
	})
	console.AddCommand("kero_drawlightmaps", "Renders lightmaps in place of diffuse textures", "kero_drawlightmaps <0|1>", func(options string) error {
		if s == nil {
			return nil
		}
		if ok := scene3.CurrentScene().Raw().TexCache.Find(scene2.LightmapTexturePath); ok == nil {
			return errors.New("kero_drawlightmaps: no lightmap in memory")
		}
		if options == "1" {
			console.SetConvarBoolean("r_drawlightmaps", true)
		} else {
			console.SetConvarBoolean("r_drawlightmaps", false)
		}
		return nil
	})

	console.AddConvarBool("hdr_enable", "Use HDR by default", true)
}

func (s *Renderer) BindSharedResources() {
	// When a new scene is loaded
	event.Get().AddListener(messages.TypeLoadingLevelParsed, func(message interface{}) {
		s.gpuScene = *scene.GpuSceneFromFrameworkScene(scene3.CurrentScene().Raw(), filesystem.Get())
	})
	// When a game is quit
	event.Get().AddListener(messages.TypeEngineDisconnect, func(e interface{}) {
		s.Cleanup()
	})

	s.bindConVars()
}

func NewRenderer() *Renderer {
	return &Renderer{}
}