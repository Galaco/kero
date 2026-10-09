package renderer

import (
	"sort"

	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/framework/graphics/adapter"
	"github.com/galaco/kero/framework/graphics/mesh"
	scene2 "github.com/galaco/kero/framework/scene"
	"github.com/galaco/kero/renderer/cache"
	"github.com/galaco/kero/renderer/scene"
	"github.com/go-gl/mathgl/mgl32"
)

// translucentItem is something translucent, drawn after everything opaque
type translucentItem struct {
	// distance is the squared distance from the camera
	distance float32
	material *cache.GpuMaterial

	// A bsp face
	face *graphics.BspFace

	// Or a static prop's sub-mesh, drawn with the instance data at instance in its batch's instance buffer
	batch    *scene.InstanceBatch
	instance int

	// Or an entity prop's sub-mesh
	gpuMesh   *adapter.GpuMesh
	subMesh   mesh.SubMesh
	transform mgl32.Mat4
}

// renderTranslucents draws translucent bsp faces and prop sub-meshes from the furthest to the nearest without writing
// depth, so each blends with everything behind it
func (s *Renderer) renderTranslucents(camera *graphics.Camera, items []translucentItem) {
	defer s.setNoCull(false)
	if len(items) == 0 {
		return
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].distance > items[j].distance
	})

	generic := s.shaderCache.Find("LightMappedGeneric")
	instanced := s.shaderCache.Find("LightMappedGenericInstanced")
	for _, shader := range []*adapter.Shader{generic, instanced} {
		shader.Bind()
		adapter.PushMat4(shader.GetUniform("projection"), 1, false, camera.ProjectionMatrix())
		adapter.PushMat4(shader.GetUniform("view"), 1, false, camera.ViewMatrix())
		adapter.PushInt32(shader.GetUniform("albedoSampler"), 0)
		adapter.PushInt32(shader.GetUniform("lightmapSampler"), 4)
		adapter.PushInt32(shader.GetUniform("hasTranslucentProperty"), 1)
	}
	adapter.PushVec3(instanced.GetUniform("cameraPosition"), camera.Transform().Translation)
	adapter.BindLightmap(s.gpuScene.GpuItemCache.Find(scene2.LightmapTexturePath))

	adapter.DisableZBufferWrite()
	defer adapter.EnableZBufferWrite()

	var boundShader *adapter.Shader
	var boundMesh adapter.GpuMesh
	bindMesh := func(gpuMesh *adapter.GpuMesh) {
		if *gpuMesh != boundMesh {
			adapter.BindMesh(gpuMesh)
			boundMesh = *gpuMesh
		}
	}
	for _, item := range items {
		shader := generic
		if item.batch != nil {
			shader = instanced
		}
		if shader != boundShader {
			shader.Bind()
			boundShader = shader
		}

		translucent := int32(0)
		if item.material.Properties.Translucent {
			translucent = 1
		}
		adapter.PushInt32(shader.GetUniform("translucent"), translucent)
		adapter.PushFloat32(shader.GetUniform("alpha"), item.material.Properties.Alpha)
		s.setNoCull(item.material.NoCull())
		pushAlphaTest(shader, item.material)
		adapter.BindTexture(item.material.Diffuse)

		switch {
		case item.face != nil:
			adapter.PushMat4(shader.GetUniform("model"), 1, false, camera.ModelMatrix())
			bindMesh(&s.gpuScene.GpuMesh)
			adapter.DrawMultiIndexedArray([]int32{int32(item.face.Length())}, []int{item.face.Offset()})
		case item.batch != nil:
			bindMesh(&item.batch.Mesh)
			adapter.SetupInstanceAttributes(item.batch.InstanceVBO, item.instance)
			adapter.DrawIndexedArrayInstanced(item.batch.IndexCount, item.batch.IndexOffset, 1)
		default:
			adapter.PushMat4(shader.GetUniform("model"), 1, false, item.transform)
			bindMesh(item.gpuMesh)
			adapter.DrawIndexedArray(item.subMesh.IndexCount, item.subMesh.IndexOffset, nil)
		}
	}

	adapter.DisableInstanceAttributes()
}

// setNoCull draws both sides of faces, or only their fronts
func (s *Renderer) setNoCull(noCull bool) {
	if noCull == s.noCull {
		return
	}
	s.noCull = noCull
	if noCull {
		adapter.DisableFaceCulling()
	} else {
		adapter.EnableBackFaceCulling()
	}
}
