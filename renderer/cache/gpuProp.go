package cache

import (
	"github.com/galaco/kero/framework/graphics/adapter"
)

// GpuProp is a model uploaded to the GPU. Its sub-meshes share Mesh.
type GpuProp struct {
	Mesh adapter.GpuMesh
	// Materials holds the material of each sub-mesh, for each skin. Skins that nothing draws are not loaded, and
	// are nil.
	Materials [][]GpuMaterial
}

// MaterialsForSkin returns the material of each sub-mesh with a skin, or with skin 0 if that skin is not loaded
func (prop *GpuProp) MaterialsForSkin(skin int) []GpuMaterial {
	if skin > 0 && skin < len(prop.Materials) && prop.Materials[skin] != nil {
		return prop.Materials[skin]
	}
	if len(prop.Materials) == 0 {
		return nil
	}
	return prop.Materials[0]
}
