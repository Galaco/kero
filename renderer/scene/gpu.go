package scene

import (
	"fmt"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/entity"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/framework/graphics/adapter"
	"github.com/galaco/kero/framework/scene"
	"github.com/galaco/kero/renderer/cache"
	"github.com/go-gl/mathgl/mgl32"
	"strings"
)

type GPUScene struct {
	Skybox *Skybox
	GpuMesh adapter.GpuMesh
	GpuItemCache cache.GpuItem
	GpuMaterialCache cache.Material
	GpuStaticProps map[string]cache.GpuProp
}

func GpuSceneFromFrameworkScene(frameworkScene *scene.StaticScene, fs fileSystem) *GPUScene {
	s := &GPUScene{
		GpuItemCache: cache.NewGpuItemCache(),
		GpuMaterialCache: cache.NewMaterialCache(),
		GpuStaticProps: map[string]cache.GpuProp{},
	}

	frameworkScene.TexCache.Add(scene.ErrorTexturePath, graphics.NewErrorTexture(scene.ErrorTexturePath))
	s.GpuItemCache.Add(scene.ErrorTexturePath, adapter.UploadTexture(frameworkScene.TexCache.Find(scene.ErrorTexturePath)))

	for key,tex := range frameworkScene.TexCache.All() {
		if key == scene.LightmapTexturePath {
			s.GpuItemCache.Add(scene.LightmapTexturePath, adapter.UploadLightmap(tex))
			tex.Release()
			continue
		}
		s.GpuItemCache.Add(key, adapter.UploadTexture(tex))
	}

	for _, mat := range frameworkScene.RawBsp.MaterialDictionary() {
		s.GpuMaterialCache.Add(strings.ToLower(mat.FilePath()), cache.NewGpuMaterial(s.GpuItemCache.Find(mat.BaseTextureName), mat))
	}

	// Finish staticprops
	for _, prop := range frameworkScene.RawBsp.StaticPropDictionary {
		gpuProp := cache.GpuProp{}
		s.GpuStaticProps[prop.Id] = cache.GpuProp{}
		for _, m := range prop.Meshes() {
			gpuProp.AddMesh(adapter.UploadMesh(m))
		}
		for _, materialPath := range prop.Materials() {
			if _, ok := frameworkScene.RawBsp.MaterialDictionary()[materialPath]; ok {
				gpuProp.AddMaterial(*s.GpuMaterialCache.Find(strings.ToLower(materialPath)))
				continue
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
			s.GpuMaterialCache.Add(strings.ToLower(mat.FilePath()), cache.NewGpuMaterial(s.GpuItemCache.Find(mat.BaseTextureName), mat))
			gpuProp.AddMaterial(*s.GpuMaterialCache.Find(strings.ToLower(materialPath)))

			s.GpuStaticProps[prop.Id] = gpuProp
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
	s.Skybox = LoadSkybox(filesystem.Get(), skyName, skyboxOrigin)
	s.GpuMesh = adapter.UploadMesh(frameworkScene.BspMesh)


	// Cleanup unneeded raw data
	for _,tex := range frameworkScene.TexCache.All() {
		tex.Release()
	}


	return s
}
