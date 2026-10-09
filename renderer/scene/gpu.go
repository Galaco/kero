package scene

import (
	"fmt"
	"strings"

	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/entity"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/framework/graphics/adapter"
	"github.com/galaco/kero/framework/graphics/mesh"
	"github.com/galaco/kero/framework/scene"
	"github.com/galaco/kero/renderer/cache"
	"github.com/go-gl/mathgl/mgl32"
)

// InstanceBatch represents a group of static props that share the same model+sub-mesh+material
// and can be rendered together using GPU instancing
type InstanceBatch struct {
	Key          string                 // Unique identifier: "modelId_meshIdx_materialPath"
	ModelId      string                 // Model identifier
	MeshIdx      int                    // Index of sub-mesh within model
	Mesh         adapter.GpuMesh        // GPU mesh handle, shared by every sub-mesh of the model
	Material     *cache.GpuMaterial     // Material the sub-mesh is drawn with
	IndexOffset  int                    // First index of the sub-mesh in Mesh
	IndexCount   int                    // Number of indices in the sub-mesh
	Props        []*graphics.StaticProp // All props in this batch (for reverse lookup)
	InstanceVBO  uint32                 // Persistent GPU buffer for instance data
	MaxInstances int                    // VBO capacity
}

type EntityPropCacheItem struct {
	Id       string
	Entities []entity.IEntity
	Prop     *mesh.Model
}

type GPUScene struct {
	Skybox                    *Skybox
	GpuMesh                   adapter.GpuMesh // Regular BSP geometry
	GpuDisplacementMesh       adapter.GpuMesh // Displacement surfaces (with blend weights)
	GpuItemCache              cache.GpuItem
	GpuMaterialCache          cache.Material
	GpuStaticProps            map[string]cache.GpuProp
	GpuRenderablePropEntities []EntityPropCacheItem
	InstanceBatches           map[string]*InstanceBatch // Pre-built instance batches for static props
	// StaticPropBatches holds the instance batches that draw each static prop, one per sub-mesh
	StaticPropBatches map[*graphics.StaticProp][]*InstanceBatch
}

func GpuSceneFromFrameworkScene(frameworkScene *scene.StaticScene, fs fileSystem) *GPUScene {
	s := &GPUScene{
		GpuItemCache:              cache.NewGpuItemCache(),
		GpuMaterialCache:          cache.NewMaterialCache(),
		GpuStaticProps:            map[string]cache.GpuProp{},
		GpuRenderablePropEntities: []EntityPropCacheItem{},
		InstanceBatches:           map[string]*InstanceBatch{},
		StaticPropBatches:         map[*graphics.StaticProp][]*InstanceBatch{},
	}

	console.PrintString(console.LevelInfo, "Submitting BSP texture data to GPU...")
	s.GpuItemCache.Add(scene.ErrorTexturePath, adapter.UploadTexture(frameworkScene.TexCache.Find(scene.ErrorTexturePath)))

	for key, tex := range frameworkScene.TexCache.All() {
		if key == scene.LightmapTexturePath {
			s.GpuItemCache.Add(scene.LightmapTexturePath, adapter.UploadLightmap(tex))
			continue
		}
		s.GpuItemCache.Add(key, adapter.UploadTexture(tex))
	}

	for _, mat := range frameworkScene.RawBsp.MaterialDictionary() {
		gpuMat := cache.NewGpuMaterial(s.GpuItemCache.Find(mat.BaseTextureName), mat)

		// If this is a blend material (WorldVertexTransition), load the second texture
		if mat.IsBlendMaterial() && mat.BaseTexture2Name != "" {
			if console.GetConvarBoolean("developer") {
				console.PrintString(console.LevelInfo, fmt.Sprintf("BSP Blend material detected: %s -> %s + %s", mat.FilePath(), mat.BaseTextureName, mat.BaseTexture2Name))
			}
			gpuMat.Diffuse2 = s.GpuItemCache.Find(mat.BaseTexture2Name)
			if gpuMat.Diffuse2 == 0 {
				// Second texture not yet loaded, load it now
				tex2, err := graphics.LoadTexture(fs, mat.BaseTexture2Name)
				if err != nil {
					console.PrintString(console.LevelWarning, fmt.Sprintf("Failed to load blend texture: %s", mat.BaseTexture2Name))
				} else {
					frameworkScene.TexCache.Add(mat.BaseTexture2Name, tex2)
					gpuMat.Diffuse2 = adapter.UploadTexture(tex2)
					s.GpuItemCache.Add(mat.BaseTexture2Name, gpuMat.Diffuse2)
					adapter.ReleaseTextureResource(tex2)
					if console.GetConvarBoolean("developer") {
						console.PrintString(console.LevelInfo, fmt.Sprintf("  Loaded blend texture %s (GPU ID: %d)", mat.BaseTexture2Name, gpuMat.Diffuse2))
					}
				}
			} else if console.GetConvarBoolean("developer") {
				console.PrintString(console.LevelInfo, fmt.Sprintf("  Using cached blend texture %s (GPU ID: %d)", mat.BaseTexture2Name, gpuMat.Diffuse2))
			}
		}

		s.GpuMaterialCache.Add(strings.ToLower(mat.FilePath()), gpuMat)
	}

	// Static props and prop entities can share models, so find every skin a model is drawn with before loading it
	skins := propSkins(frameworkScene)

	console.PrintString(console.LevelInfo, "Submitting staticprop studiomodel data to GPU...")
	// Finish staticprops
	for _, prop := range frameworkScene.RawBsp.StaticPropDictionary {
		s.LoadSingleProp(prop, skins[prop.Id], frameworkScene, fs)
	}

	console.PrintString(console.LevelInfo, "Submitting entity studiomodel data to GPU...")
	// Finish props referenced by entities
	for _, prop := range frameworkScene.RawBsp.EntityPropDictionary {
		s.LoadSingleProp(prop, skins[prop.Id], frameworkScene, fs)

		s.GpuRenderablePropEntities = append(s.GpuRenderablePropEntities, EntityPropCacheItem{
			Id:       prop.Id,
			Prop:     prop,
			Entities: make([]entity.IEntity, 0),
		})
	}

	// @TODO this will be rewritten once other systems start interacting with entities; a better shared cache is needed.
	// It's also hella slow as it doesn't make use of leaf visdata
	var entityModel string
	for _, ent := range frameworkScene.Entities {
		// @TODO Not 100% reliable; find a better way to detect if an entity is a renderable prop. "model" can also refer
		// to a BSP model so non-empty value isn't a sufficient check by itself
		if strings.HasPrefix(ent.Classname(), "prop_") {
			entityModel = ent.ValueForKey("model")
			for propIdx, r := range s.GpuRenderablePropEntities {
				if r.Id == entityModel {
					s.GpuRenderablePropEntities[propIdx].Entities = append(s.GpuRenderablePropEntities[propIdx].Entities, ent)
					break
				}
			}
		}
	}

	var worldspawn entity.IEntity
	for idx, e := range frameworkScene.Entities {
		if e.Classname() == "worldspawn" {
			worldspawn = frameworkScene.Entities[idx]
			continue
		}
	}
	skyboxOrigin := mgl32.Vec3{}
	skyName := ""
	if worldspawn != nil {
		skyboxOrigin = worldspawn.VectorForKey("origin")
		skyName = worldspawn.ValueForKey("skyname")
	}
	console.PrintString(console.LevelInfo, "Submitting bsp geometry to GPU...")
	s.Skybox = LoadSkybox(filesystem.Get(), skyName, skyboxOrigin)
	s.GpuMesh = adapter.UploadMesh(frameworkScene.BspMesh)

	console.PrintString(console.LevelInfo, "Submitting displacement geometry to GPU...")
	s.GpuDisplacementMesh = adapter.UploadMesh(frameworkScene.DisplacementBspMesh)

	// Cleanup unneeded raw data
	for _, tex := range frameworkScene.TexCache.All() {
		tex.Release()
	}

	// Build instance batches for static props
	console.PrintString(console.LevelInfo, "Building static prop instance batches...")
	s.buildInstanceBatches(frameworkScene)

	return s
}

// propSkins returns the skins each model is drawn with by static props and prop entities
func propSkins(frameworkScene *scene.StaticScene) map[string][]int {
	skins := map[string][]int{}
	for idx := range frameworkScene.RawBsp.StaticProps {
		prop := &frameworkScene.RawBsp.StaticProps[idx]
		id := prop.Model().Model.Id
		skins[id] = append(skins[id], prop.Skin())
	}
	for _, ent := range frameworkScene.Entities {
		if strings.HasPrefix(ent.Classname(), "prop_") {
			id := ent.ValueForKey("model")
			skins[id] = append(skins[id], ent.IntForKey("skin"))
		}
	}
	return skins
}

// LoadSingleProp uploads a model, and the materials of the skins it is drawn with. Skin 0 is always loaded.
func (s *GPUScene) LoadSingleProp(prop *mesh.Model, skins []int, frameworkScene *scene.StaticScene, fs fileSystem) {
	if _, ok := s.GpuStaticProps[prop.Id]; ok {
		return
	}

	gpuProp := cache.GpuProp{}
	s.GpuStaticProps[prop.Id] = gpuProp
	if len(prop.SubMeshes()) == 0 {
		return
	}
	gpuProp.Mesh = adapter.UploadMesh(prop.Mesh())
	gpuProp.Materials = make([][]cache.GpuMaterial, prop.NumSkins())
	for _, skin := range append([]int{0}, skins...) {
		skin = prop.Skin(skin)
		if gpuProp.Materials[skin] != nil {
			continue
		}
		for _, materialPath := range prop.Materials(skin) {
			gpuProp.Materials[skin] = append(gpuProp.Materials[skin], *s.loadPropMaterial(materialPath, frameworkScene, fs))
		}
	}

	s.GpuStaticProps[prop.Id] = gpuProp
}

// loadPropMaterial returns a prop material, loading it and its textures if no bsp face or other prop has
func (s *GPUScene) loadPropMaterial(materialPath string, frameworkScene *scene.StaticScene, fs fileSystem) *cache.GpuMaterial {
	if gpuMat := s.GpuMaterialCache.Find(strings.ToLower(materialPath)); gpuMat != nil {
		return gpuMat
	}

	mat, err := graphics.LoadMaterial(fs, materialPath)
	if err != nil {
		console.PrintString(console.LevelError, fmt.Sprintf("Failed to load material: %s, %s", materialPath, err.Error()))
		mat = graphics.NewMaterial(materialPath)
		mat.BaseTextureName = scene.ErrorTexturePath
	}
	if tex := frameworkScene.TexCache.Find(mat.BaseTextureName); tex == nil {
		tex, err := graphics.LoadTexture(fs, mat.BaseTextureName)
		if err != nil {
			console.PrintString(console.LevelWarning, err.Error())
			frameworkScene.TexCache.Add(mat.BaseTextureName, frameworkScene.TexCache.Find(scene.ErrorTexturePath))
			s.GpuItemCache.Add(mat.BaseTextureName, s.GpuItemCache.Find(scene.ErrorTexturePath))
		} else {
			frameworkScene.TexCache.Add(mat.BaseTextureName, tex)
			s.GpuItemCache.Add(mat.BaseTextureName, adapter.UploadTexture(tex))
			adapter.ReleaseTextureResource(tex)
		}
	}
	gpuMat := cache.NewGpuMaterial(s.GpuItemCache.Find(mat.BaseTextureName), mat)

	// If this is a blend material, load the second texture
	if mat.IsBlendMaterial() && mat.BaseTexture2Name != "" {
		if console.GetConvarBoolean("developer") {
			console.PrintString(console.LevelInfo, fmt.Sprintf("Prop Blend material detected: %s -> %s + %s", mat.FilePath(), mat.BaseTextureName, mat.BaseTexture2Name))
		}
		gpuMat.Diffuse2 = s.GpuItemCache.Find(mat.BaseTexture2Name)
		if gpuMat.Diffuse2 == 0 {
			tex2, err := graphics.LoadTexture(fs, mat.BaseTexture2Name)
			if err != nil {
				console.PrintString(console.LevelWarning, fmt.Sprintf("Failed to load blend texture: %s", mat.BaseTexture2Name))
			} else {
				frameworkScene.TexCache.Add(mat.BaseTexture2Name, tex2)
				gpuMat.Diffuse2 = adapter.UploadTexture(tex2)
				s.GpuItemCache.Add(mat.BaseTexture2Name, gpuMat.Diffuse2)
				adapter.ReleaseTextureResource(tex2)
				if console.GetConvarBoolean("developer") {
					console.PrintString(console.LevelInfo, fmt.Sprintf("  Loaded blend texture %s (GPU ID: %d)", mat.BaseTexture2Name, gpuMat.Diffuse2))
				}
			}
		} else if console.GetConvarBoolean("developer") {
			console.PrintString(console.LevelInfo, fmt.Sprintf("  Using cached blend texture %s (GPU ID: %d)", mat.BaseTexture2Name, gpuMat.Diffuse2))
		}
	}

	s.GpuMaterialCache.Add(strings.ToLower(mat.FilePath()), gpuMat)

	return gpuMat
}

// buildInstanceBatches groups all static props into instance batches for efficient rendering
// Called once at scene load time
func (s *GPUScene) buildInstanceBatches(frameworkScene *scene.StaticScene) {
	batches := make(map[string]*InstanceBatch)
	propBatches := make(map[*graphics.StaticProp][]*InstanceBatch)

	// Iterate ALL static props in the scene across all clusters
	for _, cluster := range frameworkScene.ClusterLeafs {
		for _, prop := range cluster.StaticProps {
			// A prop is listed in every cluster it touches, but belongs to its batches once
			if _, seen := propBatches[prop]; seen {
				continue
			}
			propBatches[prop] = nil

			model := prop.Model().Model
			gpuProp, ok := s.GpuStaticProps[model.Id]
			if !ok {
				continue
			}
			materials := gpuProp.MaterialsForSkin(prop.Skin())

			// Each sub-mesh in the prop might need a separate batch
			for meshIdx, subMesh := range model.SubMeshes() {
				material := &materials[meshIdx]
				key := fmt.Sprintf("%s_%d_%s", model.Id, meshIdx, material.Properties.FilePath())

				batch, exists := batches[key]
				if !exists {
					batch = &InstanceBatch{
						Key:         key,
						ModelId:     model.Id,
						MeshIdx:     meshIdx,
						Mesh:        gpuProp.Mesh,
						Material:    material,
						IndexOffset: subMesh.IndexOffset,
						IndexCount:  subMesh.IndexCount,
					}
					batches[key] = batch
				}
				batch.Props = append(batch.Props, prop)
				propBatches[prop] = append(propBatches[prop], batch)
			}
		}
	}

	// Create persistent VBOs for each batch
	const floatsPerInstance = 18 // 16 (mat4) + 2 (fade min/max)

	for _, batch := range batches {
		batch.MaxInstances = len(batch.Props)

		// Pre-allocate VBO with max capacity (but don't fill with data yet)
		batch.InstanceVBO = adapter.CreateEmptyInstanceBuffer(batch.MaxInstances, floatsPerInstance)

		if console.GetConvarBoolean("developer") {
			console.PrintString(console.LevelInfo,
				fmt.Sprintf("  Batch %s: %d instances, VBO %d",
					batch.Key, batch.MaxInstances, batch.InstanceVBO))
		}
	}

	s.InstanceBatches = batches
	s.StaticPropBatches = propBatches

	console.PrintString(console.LevelInfo,
		fmt.Sprintf("Created %d instance batches", len(batches)))
}
