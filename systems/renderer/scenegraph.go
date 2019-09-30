package renderer

import (
	"fmt"
	"github.com/galaco/kero/event"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/framework/graphics"
	graphics3d "github.com/galaco/kero/framework/graphics/3d"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/systems/renderer/cache"
	"github.com/galaco/kero/valve"
	"strings"
)

type SceneGraph struct {
	bspMesh  *graphics.Mesh
	bspFaces []valve.BspFace

	gpuMesh graphics.GpuMesh

	camera *graphics3d.Camera
}

func NewSceneGraphFromBsp(fs filesystem.FileSystem, level *valve.Bsp, materialCache *cache.Material, texCache *cache.Texture, gpuItemCache *cache.GpuItem) *SceneGraph {
	texCache.Add(cache.ErrorTexturePath, graphics.NewErrorTexture(cache.ErrorTexturePath))
	gpuItemCache.Add(cache.ErrorTexturePath, graphics.UploadTexture(texCache.Find(cache.ErrorTexturePath)))

	// load materials
	for _, mat := range level.MaterialDictionary() {
		if tex := texCache.Find(mat.BaseTextureName); tex == nil {
			tex, err := graphics.LoadTexture(fs, mat.BaseTextureName)
			if err != nil {
				event.Dispatch(messages.NewConsoleMessage(console.LevelWarning, err.Error()))
				texCache.Add(mat.BaseTextureName, texCache.Find(cache.ErrorTexturePath))
				gpuItemCache.Add(mat.BaseTextureName, gpuItemCache.Find(cache.ErrorTexturePath))
			} else {
				texCache.Add(mat.BaseTextureName, tex)
				gpuItemCache.Add(mat.BaseTextureName, graphics.UploadTexture(tex))
			}
		}
		materialCache.Add(strings.ToLower(mat.FilePath()), cache.NewGpuMaterial(gpuItemCache.Find(mat.BaseTextureName)))
	}

	// finish bsp mesh
	// Add MATERIALS TO FACES
	for _, bspFace := range level.Faces() {
		// Generate texture coordinates
		var tex *graphics.Texture
		if level.MaterialDictionary()[bspFace.Material()] == nil {
			event.Dispatch(messages.NewConsoleMessage(console.LevelWarning, fmt.Sprintf("MATERIAL: %s not found", bspFace.Material())))
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

	remappedFaces := make([]valve.BspFace, 0, 1024)
	// Kero isnt interested in tools faces (for now)
	for idx, bspFace := range level.Faces() {
		if strings.HasPrefix(strings.ToLower(bspFace.Material()), "tools") {
			continue
		}
		remappedFaces = append(remappedFaces, level.Faces()[idx])
	}

	return &SceneGraph{
		bspMesh:  level.Mesh(),
		gpuMesh:  graphics.UploadMesh(level.Mesh()),
		bspFaces: remappedFaces,
		camera:   level.Camera(),
	}
}
