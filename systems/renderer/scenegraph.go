package renderer

import (
	"github.com/galaco/kero/event"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/framework/graphics"
	graphics3d "github.com/galaco/kero/framework/graphics/3d"
	"github.com/galaco/kero/valve"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/systems/renderer/cache"
)

type SceneGraph struct {
	bspMesh  *graphics.Mesh
	bspFaces []valve.BspFace

	gpuMesh graphics.GpuMesh

	camera *graphics3d.Camera
}

func NewSceneGraphFromBsp(level *valve.Bsp, materialCache *cache.Material, texCache *cache.Texture, gpuItemCache *cache.GpuItem) *SceneGraph {
	texCache.Add(cache.ErrorTexturePath, graphics.NewErrorTexture(cache.ErrorTexturePath))
	gpuItemCache.Add(cache.ErrorTexturePath, graphics.UploadTexture(texCache.Find(cache.ErrorTexturePath)))

	// load materials
	for _, mat := range level.MaterialDictionary() {
		if tex := texCache.Find(mat.BaseTextureName); tex == nil {
			tex, err := graphics.LoadTexture(filesystem.Singleton(), mat.BaseTextureName)
			if err != nil {
				event.Singleton().Dispatch(messages.NewConsoleMessage(console.LevelWarning, err.Error()))
				texCache.Add(mat.BaseTextureName, texCache.Find(cache.ErrorTexturePath))
			} else {
				texCache.Add(mat.BaseTextureName, tex)
				gpuItemCache.Add(mat.BaseTextureName, graphics.UploadTexture(tex))
			}
		}
		materialCache.Add(mat.BaseTextureName, cache.NewGpuMaterial(gpuItemCache.Find(mat.BaseTextureName)))
	}

	// finish bsp mesh
	// Add MATERIALS TO FACES
	for _, bspFace := range level.Faces() {
		// Generate texture coordinates
		var tex *graphics.Texture
		if level.MaterialDictionary()[bspFace.Material()] == nil {
			tex = texCache.Find(cache.ErrorTexturePath)
		} else {
			if level.MaterialDictionary()[bspFace.Material()].BaseTextureName == "" {
				tex = texCache.Find(cache.ErrorTexturePath)
			} else {
				tex = texCache.Find(level.MaterialDictionary()[bspFace.Material()].BaseTextureName)
			}
		}
		level.Mesh().AddUV(
			valve.TexCoordsForFaceFromTexInfo(
				level.Mesh().Vertices()[bspFace.Offset()*3:(bspFace.Offset()*3)+(bspFace.Length()*3)],
				bspFace.TexInfo(),
				tex.Width(),
				tex.Height())...)
	}

	level.Mesh().GenerateTangents()

	return &SceneGraph{
		bspMesh:  level.Mesh(),
		gpuMesh:  graphics.UploadMesh(level.Mesh()),
		bspFaces: level.Faces(),
		camera:   level.Camera(),
	}
}
