package renderer

import (
	"errors"
	"fmt"

	"github.com/galaco/gosigl"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/ecs"
	"github.com/galaco/kero/framework/ecs/components"
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

	// Phase 4: Pure ECS
	ecsWorld *ecs.World

	dataScene *scene2.StaticScene
	gpuScene  scene.GPUScene

	activeShader *adapter.Shader
	// noCull is true while face culling is disabled for a $nocull material
	noCull bool

	// displacements groups the scene's displacements by material, to draw each material's at once
	displacements *displacementBatches

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
			translucents := s.renderBsp(s.dataScene.SkyCamera, s.dataScene.SkyboxClusterLeafs)
			s.renderDisplacements(s.dataScene.SkyCamera, s.dataScene.SkyboxClusterLeafs)
			translucents = append(translucents, s.renderStaticProps(s.dataScene.SkyCamera, s.dataScene.SkyboxClusterLeafs)...)
			s.renderTranslucents(s.dataScene.SkyCamera, translucents)
			adapter.ClearDepthBuffer()
			s.dataScene.SkyCamera.Transform().Translation = origin
		}
	}

	// Draw world. Translucent faces and props are drawn after everything opaque.
	s.startFrame(s.dataScene.Camera)
	translucents := s.renderBsp(s.dataScene.Camera, clusters)
	s.renderDisplacements(s.dataScene.Camera, clusters)
	translucents = append(translucents, s.renderStaticProps(s.dataScene.Camera, clusters)...)

	// Render entity props using ECS
	translucents = append(translucents, s.renderEntityProps()...)
	s.renderTranslucents(s.dataScene.Camera, translucents)

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
	s.displacements = newDisplacementBatches(s.dataScene.DisplacementFaces, &s.gpuScene.GpuMaterialCache)
}

func (s *Renderer) startFrame(camera *graphics.Camera) {
	projection := camera.ProjectionMatrix()
	view := camera.ViewMatrix()

	s.activeShader = s.shaderCache.Find("LightMappedGeneric")
	s.activeShader.Bind()
	adapter.PushMat4(s.activeShader.GetUniform("projection"), 1, false, projection)
	adapter.PushMat4(s.activeShader.GetUniform("view"), 1, false, view)
	adapter.PushInt32(s.activeShader.GetUniform("hasTranslucentProperty"), 0)
}

// renderBsp draws the opaque faces of clusters, and returns their translucent faces to draw later
func (s *Renderer) renderBsp(camera *graphics.Camera, clusters []*vis.ClusterLeaf) []translucentItem {
	adapter.PushMat4(s.activeShader.GetUniform("model"), 1, false, camera.ModelMatrix())
	if console.GetConvarBoolean("r_drawlightmaps") == true {
		adapter.PushInt32(s.activeShader.GetUniform("renderLightmapsAsAlbedo"), 1)
	} else {
		adapter.PushInt32(s.activeShader.GetUniform("renderLightmapsAsAlbedo"), 0)
	}

	adapter.BindMesh(&s.gpuScene.GpuMesh)
	adapter.PushInt32(s.activeShader.GetUniform("albedoSampler"), 0)
	pushLightmapUniforms(s.activeShader)
	adapter.BindLightmap(s.gpuScene.GpuItemCache.Find(scene2.LightmapTexturePath))
	var mat *cache.GpuMaterial

	materialMappedClusterFaces := vis.GroupClusterFacesByMaterial(clusters)
	cameraPosition := camera.Transform().Translation
	translucents := make([]translucentItem, 0)

	for clusterFaceMaterial, faces := range materialMappedClusterFaces {
		mat = s.gpuScene.GpuMaterialCache.Find(clusterFaceMaterial)

		if mat.Properties.Skip {
			continue
		}

		if mat.IsTranslucent() {
			for _, face := range faces {
				translucents = append(translucents, translucentItem{
					distance: face.Center().Sub(cameraPosition).LenSqr(),
					material: mat,
					face:     face,
				})
			}
			continue
		}
		s.RenderBSPMaterial(mat, faces)
	}
	s.setNoCull(false)

	return translucents
}

func (s *Renderer) RenderBSPMaterial(mat *cache.GpuMaterial, faces []*graphics.BspFace) {
	// Build offset/count arrays for multi-draw (no index array allocation/upload needed)
	counts := make([]int32, len(faces))
	offsets := make([]int, len(faces))
	for i, face := range faces {
		counts[i] = int32(face.Length())
		offsets[i] = face.Offset()
	}

	s.setNoCull(mat.NoCull())
	pushAlphaTest(s.activeShader, mat)
	pushBumpmap(s.activeShader, mat)
	adapter.BindTexture(mat.Diffuse)
	adapter.DrawMultiIndexedArray(counts, offsets)
	if err := adapter.GpuError(); err != nil {
		console.PrintString(console.LevelError, err.Error())
	}
}

// pushAlphaTest tells a shader whether to discard the parts of a material's base texture below its alpha test
// reference
func pushAlphaTest(shader *adapter.Shader, mat *cache.GpuMaterial) {
	alphaTest, reference := mat.AlphaTest()
	enabled := int32(0)
	if alphaTest {
		enabled = 1
	}
	adapter.PushInt32(shader.GetUniform("alphaTest"), enabled)
	adapter.PushFloat32(shader.GetUniform("alphaTestReference"), reference)
}

// pushLightmapUniforms tells a shader where to find lightmaps and normal maps
func pushLightmapUniforms(shader *adapter.Shader) {
	adapter.PushInt32(shader.GetUniform("lightmapSampler"), 4)
	adapter.PushInt32(shader.GetUniform("bumpmapSampler"), 2)
	adapter.PushFloat32(shader.GetUniform("lightmapPageHeight"), 1/float32(graphics.LightmapPages))
}

// pushBumpmap binds a material's normal map, and tells a shader how to light the material with it
func pushBumpmap(shader *adapter.Shader, mat *cache.GpuMaterial) {
	if mat.Bumpmap != 0 {
		gosigl.BindTexture2D(gosigl.TextureSlot(2), gosigl.TextureBindingId(mat.Bumpmap))
	}
	adapter.PushInt32(shader.GetUniform("bumpmap"), mat.BumpmapMode())
}

// renderDisplacements draws the displacements in clusters that camera sees, all of a material's at once
func (s *Renderer) renderDisplacements(camera *graphics.Camera, clusters []*vis.ClusterLeaf) {
	s.displacements.collect(clusters, graphics.FrustumFromCamera(camera))

	// Every displacement is in the displacement mesh, and both displacement shaders read the lightmap from the same
	// texture slot
	adapter.BindMesh(&s.gpuScene.GpuDisplacementMesh)
	adapter.BindLightmap(s.gpuScene.GpuItemCache.Find(scene2.LightmapTexturePath))
	adapter.PushInt32(s.activeShader.GetUniform("albedoSampler"), 0)
	pushLightmapUniforms(s.activeShader)

	var blendShader *adapter.Shader
	for _, batch := range s.displacements.batches {
		if len(batch.offsets) == 0 {
			continue
		}

		s.setNoCull(batch.material.NoCull())
		if !batch.blend {
			pushAlphaTest(s.activeShader, batch.material)
			pushBumpmap(s.activeShader, batch.material)
			adapter.BindTexture(batch.material.Diffuse)
		} else {
			// Blended materials are after the others, so the shader is switched once
			if blendShader == nil {
				blendShader = s.bindWorldVertexTransition(camera)
				if blendShader == nil {
					break
				}
			}
			pushAlphaTest(blendShader, batch.material)
			pushBumpmap(blendShader, batch.material)
			adapter.BindTexture(batch.material.Diffuse)
			second := batch.material.Diffuse2
			if second == 0 {
				second = batch.material.Diffuse
			}
			gosigl.BindTexture2D(gosigl.TextureSlot(1), gosigl.TextureBindingId(second))
			hasBumpmap2 := int32(0)
			if batch.material.Bumpmap2 != 0 {
				gosigl.BindTexture2D(gosigl.TextureSlot(3), gosigl.TextureBindingId(batch.material.Bumpmap2))
				hasBumpmap2 = 1
			}
			adapter.PushInt32(blendShader.GetUniform("hasBumpmap2"), hasBumpmap2)
		}

		adapter.DrawMultiIndexedArray(batch.counts, batch.offsets)
		if err := adapter.GpuError(); err != nil {
			console.PrintString(console.LevelError, err.Error())
		}
	}

	s.setNoCull(false)
	if blendShader != nil {
		// Restore LightMappedGeneric shader for subsequent rendering
		s.activeShader.Bind()
	}
}

// bindWorldVertexTransition binds the shader that blends displacements between two textures, to draw for camera
func (s *Renderer) bindWorldVertexTransition(camera *graphics.Camera) *adapter.Shader {
	shader := s.shaderCache.Find("WorldVertexTransition")
	if shader == nil {
		console.PrintString(console.LevelError, "WorldVertexTransition shader not found")
		return nil
	}

	shader.Bind()
	adapter.PushMat4(shader.GetUniform("projection"), 1, false, camera.ProjectionMatrix())
	adapter.PushMat4(shader.GetUniform("view"), 1, false, camera.ViewMatrix())
	adapter.PushMat4(shader.GetUniform("model"), 1, false, camera.ModelMatrix())
	adapter.PushInt32(shader.GetUniform("basetextureSampler"), 0)
	adapter.PushInt32(shader.GetUniform("basetexture2Sampler"), 1)
	adapter.PushInt32(shader.GetUniform("bumpmap2Sampler"), 3)
	pushLightmapUniforms(shader)

	if console.GetConvarBoolean("r_drawlightmaps") == true {
		adapter.PushInt32(shader.GetUniform("renderLightmapsAsAlbedo"), 1)
	} else {
		adapter.PushInt32(shader.GetUniform("renderLightmapsAsAlbedo"), 0)
	}

	adapter.PushInt32(shader.GetUniform("hasTranslucentProperty"), 0)
	adapter.PushFloat32(shader.GetUniform("alpha"), 0)
	adapter.PushInt32(shader.GetUniform("translucent"), 0)

	return shader
}

// renderStaticProps draws the opaque sub-meshes of static props in clusters, and returns their translucent sub-meshes
// to draw later
func (s *Renderer) renderStaticProps(camera *graphics.Camera, clusters []*vis.ClusterLeaf) []translucentItem {
	viewFrustum := graphics.FrustumFromCamera(camera)
	translucents := make([]translucentItem, 0)

	// Build list of visible instance data per batch
	// Value = flat array of instance data (mat4 + vec2 fade)
	batchVisibleData := make(map[*scene.InstanceBatch][]float32)
	// A prop is listed in every cluster it touches, but is drawn once
	visited := make(map[*graphics.StaticProp]bool)

	// Iterate visible clusters and props to determine which instances are visible
	for _, cluster := range clusters {
		for _, prop := range cluster.StaticProps {
			if visited[prop] {
				continue
			}
			visited[prop] = true

			// Props beyond their fade distance are fully faded out. The shader fades by the same distance.
			if prop.FadeMaxDistance() > 0 && prop.Transform.Translation.Sub(camera.Transform().Translation).Len() >= prop.FadeMaxDistance() {
				continue
			}

			// Per-prop frustum culling using accurate transformed bounds
			propMins, propMaxs := prop.GetTransformedBounds()
			if !viewFrustum.IsCuboidInFrustum(propMins, propMaxs) {
				continue
			}

			// Pack instance data: mat4 (16 floats) + vec2 fade (2 floats) = 18 floats
			transform := prop.Transform.TransformationMatrix()
			instanceData := make([]float32, 18)
			copy(instanceData, transform[:])
			instanceData[16] = prop.FadeMinDistance()
			instanceData[17] = prop.FadeMaxDistance()

			// Add visible prop to the batch of each of its sub-meshes. Translucent sub-meshes are drawn one at a time
			// later, with the instance data uploaded here.
			propCenter := propMins.Add(propMaxs).Mul(0.5)
			for _, batch := range s.gpuScene.StaticPropBatches[prop] {
				if batch.Material.IsTranslucent() {
					translucents = append(translucents, translucentItem{
						distance: propCenter.Sub(camera.Transform().Translation).LenSqr(),
						material: batch.Material,
						batch:    batch,
						instance: len(batchVisibleData[batch]) / 18,
					})
				}
				batchVisibleData[batch] = append(batchVisibleData[batch], instanceData...)
			}
		}
	}

	// Switch to instanced shader
	instancedShader := s.shaderCache.Find("LightMappedGenericInstanced")
	if instancedShader == nil {
		console.PrintString(console.LevelError, "Failed to find LightMappedGenericInstanced shader")
		return nil
	}
	instancedShader.Bind()
	adapter.PushMat4(instancedShader.GetUniform("projection"), 1, false, camera.ProjectionMatrix())
	adapter.PushMat4(instancedShader.GetUniform("view"), 1, false, camera.ViewMatrix())
	adapter.PushVec3(instancedShader.GetUniform("cameraPosition"), camera.Transform().Translation)
	adapter.PushInt32(instancedShader.GetUniform("albedoSampler"), 0)
	adapter.PushInt32(instancedShader.GetUniform("lightmapSampler"), 4)
	adapter.PushInt32(instancedShader.GetUniform("hasTranslucentProperty"), 0)
	adapter.BindLightmap(s.gpuScene.GpuItemCache.Find(scene2.LightmapTexturePath))

	if err := adapter.GpuError(); err != nil {
		console.PrintString(console.LevelError, fmt.Sprintf("GL error after instanced shader setup: %s", err.Error()))
	}

	// Render each batch that has visible instances
	for batch, data := range batchVisibleData {
		instanceCount := len(data) / 18 // 18 floats per instance

		// Skip if no instances (shouldn't happen, but safety check)
		if instanceCount == 0 {
			continue
		}

		// Update persistent VBO with visible instances only
		adapter.UpdateInstanceBuffer(batch.InstanceVBO, data)
		if batch.Material.IsTranslucent() {
			continue
		}

		// Bind mesh and material
		adapter.BindMesh(&batch.Mesh)
		s.setNoCull(batch.Material.NoCull())
		pushAlphaTest(instancedShader, batch.Material)
		adapter.BindTexture(batch.Material.Diffuse)
		adapter.SetupInstanceAttributes(batch.InstanceVBO, 0)

		// Draw all visible instances with one call!
		adapter.DrawIndexedArrayInstanced(batch.IndexCount, batch.IndexOffset, instanceCount)

		// Check for GL errors after draw
		if err := adapter.GpuError(); err != nil {
			console.PrintString(console.LevelError, fmt.Sprintf("GL error drawing batch %s: %s", batch.Key, err.Error()))
		}
	}

	// IMPORTANT: Disable instance attributes before switching to non-instanced rendering
	// Entity props reuse the same mesh VAOs but don't provide instance data
	adapter.DisableInstanceAttributes()
	s.setNoCull(false)

	return translucents
}

// renderEntityProps renders the opaque sub-meshes of entities using ECS queries, and returns their translucent
// sub-meshes to draw later
func (s *Renderer) renderEntityProps() []translucentItem {
	// Query entities with Transform + Model components
	query := s.ecsWorld.Query().
		With(ecs.ComponentTypeTransform).
		With(ecs.ComponentTypeModel).
		Build()

	entities := query.Entities()
	if len(entities) == 0 {
		return nil
	}
	translucents := make([]translucentItem, 0)

	// Switch back to non-instanced shader for entity rendering
	s.activeShader = s.shaderCache.Find("LightMappedGeneric")
	s.activeShader.Bind()

	// Re-set uniforms for non-instanced shader (uniforms are per-program)
	adapter.PushMat4(s.activeShader.GetUniform("projection"), 1, false, s.dataScene.Camera.ProjectionMatrix())
	adapter.PushMat4(s.activeShader.GetUniform("view"), 1, false, s.dataScene.Camera.ViewMatrix())
	adapter.PushInt32(s.activeShader.GetUniform("albedoSampler"), 0)
	adapter.PushInt32(s.activeShader.GetUniform("lightmapSampler"), 4)
	adapter.PushInt32(s.activeShader.GetUniform("hasTranslucentProperty"), 0)
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
		if gpuProp, ok := s.gpuScene.GpuStaticProps[modelId]; ok && gpuProp.Mesh != nil {
			adapter.BindMesh(&gpuProp.Mesh)
			materials := gpuProp.MaterialsForSkin(model.Skin)
			for idx, subMesh := range modelInstance.Model.SubMeshes() {
				if materials[idx].IsTranslucent() {
					translucents = append(translucents, translucentItem{
						distance:  transform.Position.Sub(s.dataScene.Camera.Transform().Translation).LenSqr(),
						material:  &materials[idx],
						gpuMesh:   &gpuProp.Mesh,
						subMesh:   subMesh,
						transform: transformMatrix,
					})
					continue
				}
				s.setNoCull(materials[idx].NoCull())
				pushAlphaTest(s.activeShader, &materials[idx])
				adapter.BindTexture(materials[idx].Diffuse)
				adapter.DrawIndexedArray(subMesh.IndexCount, subMesh.IndexOffset, nil)
			}
		}
	}
	s.setNoCull(false)

	return translucents
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
	adapter.DrawArray(0, len(skybox.SkyMesh.Vertices())/3)
}

func (s *Renderer) Cleanup() {
	// Release GPU resources

	// Delete instance buffers
	for _, batch := range s.gpuScene.InstanceBatches {
		adapter.DeleteInstanceBuffer(batch.InstanceVBO)
	}

	for _, s := range s.gpuScene.GpuStaticProps {
		if s.Mesh != nil {
			adapter.DeleteMeshResource(s.Mesh)
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
func NewRenderer(eventBus *event.Dispatcher, fileSystem filesystem.FileSystem, ecsWorld *ecs.World) *Renderer {
	return &Renderer{
		eventBus:   eventBus,
		fileSystem: fileSystem,
		ecsWorld:   ecsWorld,
	}
}
