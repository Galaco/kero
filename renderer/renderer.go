package renderer

import (
	"errors"
	"fmt"
	"math"

	"github.com/galaco/gosigl"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/ecs"
	"github.com/galaco/kero/framework/ecs/components"
	"github.com/galaco/kero/framework/ecs/legacy"
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/framework/graphics/adapter"
	gfxdebug "github.com/galaco/kero/framework/graphics/debug"
	"github.com/galaco/kero/framework/graphics/mesh"
	scene2 "github.com/galaco/kero/framework/scene"
	"github.com/galaco/kero/framework/scene/vis"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/renderer/cache"
	renderdebug "github.com/galaco/kero/renderer/debug"
	"github.com/galaco/kero/renderer/scene"
	"github.com/galaco/kero/renderer/shaders"
	"github.com/galaco/kero/utils"
	"github.com/go-gl/mathgl/mgl32"
)

type Renderer struct {
	eventBus    *event.Dispatcher
	fileSystem  filesystem.FileSystem
	shaderCache *cache.Shader

	// Phase 4: ECS integration
	ecsWorld     *ecs.World
	legacyBridge *legacy.Bridge

	dataScene *scene2.StaticScene
	gpuScene  scene.GPUScene

	activeShader *adapter.Shader

	// Debug rendering system
	debugRenderer *renderdebug.DebugRenderer
	debugBuffer   *gfxdebug.DebugDrawBuffer

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

	// Initialize debug rendering system
	s.debugBuffer = gfxdebug.NewDebugDrawBuffer()
	debugShader := s.shaderCache.Find("Debug")
	if debugShader != nil {
		s.debugRenderer = renderdebug.NewDebugRenderer(debugShader, s.debugBuffer)
	}

	// Register typed event listeners (Phase 3)
	event.RegisterTypedEvent(s.eventBus, s.onLoadingLevelParsedTyped)
	event.RegisterTypedEvent(s.eventBus, func(e messages.EngineDisconnectEvent) {
		s.Cleanup()
	})
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
			origin := s.dataScene.SkyCamera.Transform().Translation
			s.dataScene.SkyCamera.Transform().Orientation = s.dataScene.Camera.Transform().Orientation
			s.dataScene.SkyCamera.Transform().Translation = s.dataScene.SkyCamera.Transform().Translation.Add(s.dataScene.Camera.Transform().Translation.Mul(1 / s.dataScene.SkyCamera.Transform().Scale.X()))
			s.dataScene.SkyCamera.Update(0)
			s.startFrame(s.dataScene.SkyCamera)
			s.renderBsp(s.dataScene.SkyCamera, s.dataScene.SkyboxClusterLeafs)
			s.renderDisplacements(s.dataScene.DisplacementFaces)
			s.renderStaticProps(s.dataScene.SkyCamera, s.dataScene.SkyboxClusterLeafs)
			adapter.ClearDepthBuffer()
			s.dataScene.SkyCamera.Transform().Translation = origin
		}
	}

	// Draw world
	s.startFrame(s.dataScene.Camera)
	s.renderBsp(s.dataScene.Camera, clusters)
	s.renderDisplacements(s.dataScene.DisplacementFaces)
	s.renderStaticProps(s.dataScene.Camera, clusters)

	// Render entity props using ECS
	s.renderEntityProps()

	// Render debug primitives using new debug system
	if s.debugRenderer != nil {
		s.DrawDebug()
		projection := s.dataScene.Camera.ProjectionMatrix()
		view := s.dataScene.Camera.ViewMatrix()
		s.debugRenderer.Render(projection, view)
		s.debugBuffer.Clear()
	}
}

func (s *Renderer) DrawDebug() {
	if s.debugBuffer == nil {
		return
	}

	// Leafvis debug visualization
	switch console.GetConvarInt("mat_leafvis") {
	case 1:
		// All cluster leafs
		for _, l := range s.dataScene.ClusterLeafs {
			verts := s.convertCuboidToVec3(mesh.NewCuboidFromMinMaxs(
				mgl32.Vec3{l.Mins.X(), l.Mins.Y(), l.Mins.Z()},
				mgl32.Vec3{l.Maxs.X(), l.Maxs.Y(), l.Maxs.Z()},
			))
			s.debugBuffer.AddLines(verts, mgl32.Vec3{0, 1, 0}, mgl32.Ident4())
		}
	case 2:
		// Current leaf only
		if s.dataScene.CurrentLeaf != nil {
			verts := s.convertCuboidToVec3(mesh.NewCuboidFromMinMaxs(
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
			))
			s.debugBuffer.AddLines(verts, mgl32.Vec3{0, 1, 0}, mgl32.Ident4())
		}
	case 3:
		// Visible cluster leafs
		for _, l := range s.dataScene.VisibleClusterLeafs {
			verts := s.convertCuboidToVec3(mesh.NewCuboidFromMinMaxs(
				mgl32.Vec3{l.Mins.X(), l.Mins.Y(), l.Mins.Z()},
				mgl32.Vec3{l.Maxs.X(), l.Maxs.Y(), l.Maxs.Z()},
			))
			s.debugBuffer.AddLines(verts, mgl32.Vec3{0, 1, 0}, mgl32.Ident4())
		}
	}
}

// convertCuboidToVec3 converts a cuboid mesh to Vec3 vertices for debug rendering
func (s *Renderer) convertCuboidToVec3(cuboid *mesh.Cube) []mgl32.Vec3 {
	verts := cuboid.Vertices()
	result := make([]mgl32.Vec3, 0, len(verts)/3)
	for i := 0; i < len(verts); i += 3 {
		result = append(result, mgl32.Vec3{verts[i], verts[i+1], verts[i+2]})
	}
	return result
}

func (s *Renderer) FinishFrame() {
	adapter.ClearColor(0.25, 0.25, 0.25, 1)
	adapter.ClearAll()
}

func (s *Renderer) onLoadingLevelParsedTyped(e messages.LoadingLevelParsedEvent) {
	s.dataScene = e.Level
	s.gpuScene = *scene.GpuSceneFromFrameworkScene(s.dataScene, s.fileSystem)
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
	// Build offset/count arrays for multi-draw (no index array allocation/upload needed)
	counts := make([]int32, len(faces))
	offsets := make([]int, len(faces))
	for i, face := range faces {
		counts[i] = int32(face.Length())
		offsets[i] = face.Offset()
	}

	adapter.BindTexture(mat.Diffuse)
	adapter.DrawMultiIndexedArray(counts, offsets)
	if err := adapter.GpuError(); err != nil {
		console.PrintString(console.LevelError, err.Error())
	}
}

func (s *Renderer) renderDisplacements(displacements []*graphics.BspFace) {
	var mat *cache.GpuMaterial
	var currentShader *adapter.Shader

	// Group displacements by shader type (blend vs non-blend)
	blendDisplacements := make([]*graphics.BspFace, 0)
	regularDisplacements := make([]*graphics.BspFace, 0)

	for _, displacement := range displacements {
		mat = s.gpuScene.GpuMaterialCache.Find(displacement.Material())
		if mat != nil && mat.Properties.IsBlendMaterial() {
			blendDisplacements = append(blendDisplacements, displacement)
		} else {
			regularDisplacements = append(regularDisplacements, displacement)
		}
	}

	// ALL displacements are in the displacement mesh, so bind it first
	adapter.BindMesh(&s.gpuScene.GpuDisplacementMesh)

	// Render regular displacements with LightMappedGeneric shader (blend weights ignored)
	if len(regularDisplacements) > 0 {
		// LightMappedGeneric is already bound from startFrame, just ensure uniforms are set
		adapter.PushInt32(s.activeShader.GetUniform("albedoSampler"), 0)
		adapter.PushInt32(s.activeShader.GetUniform("lightmapSampler"), 4)
		adapter.BindLightmap(s.gpuScene.GpuItemCache.Find(scene2.LightmapTexturePath))

		for _, displacement := range regularDisplacements {
			mat = s.gpuScene.GpuMaterialCache.Find(displacement.Material())
			adapter.DrawFace(displacement.Offset(), displacement.Length(), mat.Diffuse)
			if err := adapter.GpuError(); err != nil {
				console.PrintString(console.LevelError, err.Error())
			}
		}
	}

	// Render blend displacements with WorldVertexTransition
	if len(blendDisplacements) > 0 {
		currentShader = s.shaderCache.Find("WorldVertexTransition")
		if currentShader == nil {
			console.PrintString(console.LevelError, "WorldVertexTransition shader not found")
			return
		}

		currentShader.Bind()
		// Displacement mesh already bound above, no need to re-bind

		adapter.PushMat4(currentShader.GetUniform("projection"), 1, false, s.dataScene.Camera.ProjectionMatrix())
		adapter.PushMat4(currentShader.GetUniform("view"), 1, false, s.dataScene.Camera.ViewMatrix())
		adapter.PushMat4(currentShader.GetUniform("model"), 1, false, s.dataScene.Camera.ModelMatrix())
		adapter.PushInt32(currentShader.GetUniform("basetextureSampler"), 0)
		adapter.PushInt32(currentShader.GetUniform("basetexture2Sampler"), 1)
		adapter.PushInt32(currentShader.GetUniform("lightmapSampler"), 2)

		if console.GetConvarBoolean("r_drawlightmaps") == true {
			adapter.PushInt32(currentShader.GetUniform("renderLightmapsAsAlbedo"), 1)
		} else {
			adapter.PushInt32(currentShader.GetUniform("renderLightmapsAsAlbedo"), 0)
		}

		adapter.PushInt32(currentShader.GetUniform("hasTranslucentProperty"), 0)
		adapter.PushFloat32(currentShader.GetUniform("alpha"), 0)
		adapter.PushInt32(currentShader.GetUniform("translucent"), 0)

		for _, displacement := range blendDisplacements {
			mat = s.gpuScene.GpuMaterialCache.Find(displacement.Material())

			// Bind first texture to texture unit 0
			adapter.BindTexture(mat.Diffuse)

			// Bind second texture to texture unit 1
			if mat.Diffuse2 != 0 {
				gosigl.BindTexture2D(gosigl.TextureSlot(1), gosigl.TextureBindingId(mat.Diffuse2))
			} else if console.GetConvarBoolean("developer") {
				console.PrintString(console.LevelWarning, "  Blend displacement has no second texture!")
			}

			// Bind lightmap to texture unit 2
			gosigl.BindTexture2D(gosigl.TextureSlot(2), gosigl.TextureBindingId(s.gpuScene.GpuItemCache.Find(scene2.LightmapTexturePath)))

			adapter.DrawArray(displacement.Offset(), displacement.Length())
			if err := adapter.GpuError(); err != nil {
				console.PrintString(console.LevelError, err.Error())
			}
		}

		// Restore LightMappedGeneric shader for subsequent rendering
		s.activeShader.Bind()
	}
}

func (s *Renderer) renderStaticProps(camera *graphics.Camera, clusters []*vis.ClusterLeaf) {
	viewFrustum := graphics.FrustumFromCamera(camera)

	// Build list of visible instance data per batch
	// Key = batch key, Value = flat array of instance data (mat4 + vec2 fade)
	batchVisibleData := make(map[string][]float32)

	// Iterate visible clusters and props to determine which instances are visible
	for _, cluster := range clusters {
		distToCluster := math.Pow(float64(cluster.Origin.X()-camera.Transform().Translation.X()), 2) +
			math.Pow(float64(cluster.Origin.Y()-camera.Transform().Translation.Y()), 2) +
			math.Pow(float64(cluster.Origin.Z()-camera.Transform().Translation.Z()), 2)

		for _, prop := range cluster.StaticProps {
			// Fade distance check
			if prop.FadeMaxDistance() > 0 && distToCluster >= math.Pow(float64(prop.FadeMaxDistance()), 2) {
				continue
			}

			// Per-prop frustum culling using accurate transformed bounds
			propMins, propMaxs := prop.GetTransformedBounds()
			if !viewFrustum.IsCuboidInFrustum(propMins, propMaxs) {
				continue
			}

			// Add visible prop to its batch(es)
			gpuProp, ok := s.gpuScene.GpuStaticProps[prop.Model().Model.Id]
			if !ok {
				continue
			}

			for meshIdx := range gpuProp.Id {
				materialHash := gpuProp.Material[meshIdx].Diffuse
				batchKey := fmt.Sprintf("%s_%d_%d", prop.Model().Model.Id, meshIdx, materialHash)

				// Pack instance data: mat4 (16 floats) + vec2 fade (2 floats) = 18 floats
				transform := prop.Transform.TransformationMatrix()
				instanceData := make([]float32, 18)

				// Copy matrix (16 floats, column-major order)
				for i := 0; i < 16; i++ {
					instanceData[i] = transform[i]
				}

				// Add fade distances (2 floats)
				instanceData[16] = prop.FadeMinDistance()
				instanceData[17] = prop.FadeMaxDistance()

				// Append to batch's data array
				batchVisibleData[batchKey] = append(batchVisibleData[batchKey], instanceData...)
			}
		}
	}

	// Switch to instanced shader
	instancedShader := s.shaderCache.Find("LightMappedGenericInstanced")
	if instancedShader == nil {
		console.PrintString(console.LevelError, "Failed to find LightMappedGenericInstanced shader")
		return
	}
	instancedShader.Bind()
	adapter.PushMat4(instancedShader.GetUniform("projection"), 1, false, camera.ProjectionMatrix())
	adapter.PushMat4(instancedShader.GetUniform("view"), 1, false, camera.ViewMatrix())
	adapter.PushVec3(instancedShader.GetUniform("cameraPosition"), camera.Transform().Translation)
	adapter.PushInt32(instancedShader.GetUniform("albedoSampler"), 0)
	adapter.PushInt32(instancedShader.GetUniform("lightmapSampler"), 4)
	adapter.BindLightmap(s.gpuScene.GpuItemCache.Find(scene2.LightmapTexturePath))

	if err := adapter.GpuError(); err != nil {
		console.PrintString(console.LevelError, fmt.Sprintf("GL error after instanced shader setup: %s", err.Error()))
	}

	// Render each batch that has visible instances
	for batchKey, data := range batchVisibleData {
		batch, exists := s.gpuScene.InstanceBatches[batchKey]
		if !exists {
			continue
		}

		instanceCount := len(data) / 18 // 18 floats per instance

		// Skip if no instances (shouldn't happen, but safety check)
		if instanceCount == 0 {
			continue
		}

		// Update persistent VBO with visible instances only
		adapter.UpdateInstanceBuffer(batch.InstanceVBO, data)

		// Bind mesh and material
		adapter.BindMesh(&batch.Mesh)
		adapter.BindTexture(batch.Material)
		adapter.SetupInstanceAttributes(batch.InstanceVBO)

		// Draw all visible instances with one call!
		adapter.DrawIndexedArrayInstanced(batch.IndexCount, instanceCount)

		// Check for GL errors after draw
		if err := adapter.GpuError(); err != nil {
			console.PrintString(console.LevelError, fmt.Sprintf("GL error drawing batch %s: %s", batchKey, err.Error()))
		}
	}

	// IMPORTANT: Disable instance attributes before switching to non-instanced rendering
	// Entity props reuse the same mesh VAOs but don't provide instance data
	adapter.DisableInstanceAttributes()
}

// renderEntityProps renders entities using ECS queries
func (s *Renderer) renderEntityProps() {
	// Query entities with Transform + Model components
	query := s.ecsWorld.Query().
		With(ecs.ComponentTypeTransform).
		With(ecs.ComponentTypeModel).
		Build()

	entities := query.Entities()
	if len(entities) == 0 {
		return
	}

	// Switch back to non-instanced shader for entity rendering
	s.activeShader = s.shaderCache.Find("LightMappedGeneric")
	s.activeShader.Bind()

	// Re-set uniforms for non-instanced shader (uniforms are per-program)
	adapter.PushMat4(s.activeShader.GetUniform("projection"), 1, false, s.dataScene.Camera.ProjectionMatrix())
	adapter.PushMat4(s.activeShader.GetUniform("view"), 1, false, s.dataScene.Camera.ViewMatrix())
	adapter.PushInt32(s.activeShader.GetUniform("albedoSampler"), 0)
	adapter.PushInt32(s.activeShader.GetUniform("lightmapSampler"), 4)
	adapter.BindLightmap(s.gpuScene.GpuItemCache.Find(scene2.LightmapTexturePath))

	// Render each entity
	for _, entity := range entities {
		transform, _ := ecs.GetComponent[components.Transform](s.ecsWorld, entity)
		model, _ := ecs.GetComponent[components.Model](s.ecsWorld, entity)

		// Skip invisible or uninitialized models
		if !model.Visible || !model.HasModelInstance() {
			continue
		}

		// Phase 2: Get ModelInstance directly from component (no bridge!)
		modelInstance := model.GetModelInstance().(*mesh.ModelInstance)

		// Create transformation matrix from ECS transform
		transformMatrix := mgl32.Translate3D(transform.Position.X(), transform.Position.Y(), transform.Position.Z()).
			Mul4(transform.Orientation.Mat4()).
			Mul4(mgl32.Scale3D(transform.Scale.X(), transform.Scale.Y(), transform.Scale.Z()))

		adapter.PushMat4(s.activeShader.GetUniform("model"), 1, false, transformMatrix)

		// Render using cached GPU resources
		modelId := modelInstance.Model.Id
		if gpuProp, ok := s.gpuScene.GpuStaticProps[modelId]; ok {
			for idx := range gpuProp.Id {
				adapter.BindMesh(&gpuProp.Id[idx])
				adapter.BindTexture(gpuProp.Material[idx].Diffuse)
				adapter.DrawIndexedArray(len(modelInstance.Model.Meshes()[idx].Indices()), 0, nil)
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
	skyboxTransform.Translation = s.dataScene.Camera.Transform().Translation

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

func (s *Renderer) Cleanup() {
	// Release GPU resources

	// Delete instance buffers
	for _, batch := range s.gpuScene.InstanceBatches {
		adapter.DeleteInstanceBuffer(batch.InstanceVBO)
	}

	for _, s := range s.gpuScene.GpuStaticProps {
		for _, id := range s.Id {
			adapter.DeleteMeshResource(id)
		}
	}

	for _, id := range s.gpuScene.GpuItemCache.All() {
		adapter.DeleteTextureResource(id)
	}

	// Cleanup debug renderer
	if s.debugRenderer != nil {
		s.debugRenderer.Cleanup()
	}

	s.gpuScene = scene.GPUScene{}
	s.dataScene = nil
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

// GetDebugBuffer returns the debug draw buffer for external systems to populate
func (s *Renderer) GetDebugBuffer() *gfxdebug.DebugDrawBuffer {
	return s.debugBuffer
}

// NewRenderer creates a new renderer with explicit dependencies
func NewRenderer(eventBus *event.Dispatcher, fileSystem filesystem.FileSystem, ecsWorld *ecs.World, legacyBridge *legacy.Bridge) *Renderer {
	return &Renderer{
		eventBus:     eventBus,
		fileSystem:   fileSystem,
		ecsWorld:     ecsWorld,
		legacyBridge: legacyBridge,
	}
}
