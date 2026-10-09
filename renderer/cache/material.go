package cache

import (
	"github.com/galaco/kero/framework/graphics"
)

const (
	ErrorMaterialPath = "materials/error.vmt"
)

type GpuMaterial struct {
	Diffuse  uint32
	Diffuse2 uint32 // Second texture for blend materials (WorldVertexTransition)
	// Bumpmap and Bumpmap2 are the normal maps of the material's textures, or 0 for none
	Bumpmap    uint32
	Bumpmap2   uint32
	Properties *graphics.Material
}

func NewGpuMaterial(diffuse uint32, mat *graphics.Material) *GpuMaterial {
	return &GpuMaterial{
		Diffuse:    diffuse,
		Diffuse2:   0, // Will be set if this is a blend material
		Properties: mat,
	}
}

// IsTranslucent returns true if the material must be drawn after everything opaque
func (mat *GpuMaterial) IsTranslucent() bool {
	return mat.Properties != nil && mat.Properties.IsTranslucent()
}

// NoCull returns true if both sides of faces with this material are drawn
func (mat *GpuMaterial) NoCull() bool {
	return mat.Properties != nil && mat.Properties.NoCull
}

// BumpmapMode tells a shader how to light the material with its normal maps: 0 without them, 1 from the normal of a
// normal map, or 2 from a self-shadowing bump map ($ssbump)
func (mat *GpuMaterial) BumpmapMode() int32 {
	switch {
	case mat.Bumpmap == 0:
		return 0
	case mat.Properties != nil && mat.Properties.SSBump:
		return 2
	default:
		return 1
	}
}

// AlphaTest returns whether parts of the material's base texture are discarded, and the alpha they are discarded below
func (mat *GpuMaterial) AlphaTest() (bool, float32) {
	if mat.Properties == nil || !mat.Properties.AlphaTest {
		return false, 0
	}
	return true, mat.Properties.AlphaTestReference
}

type Material struct {
	items map[string]*GpuMaterial
}

// Add
func (cache *Material) Add(name string, item *GpuMaterial) {
	cache.items[name] = item
}

// Find
func (cache *Material) Find(name string) *GpuMaterial {
	return cache.items[name]
}

// NewTextureCache
func NewMaterialCache() Material {
	return Material{
		items: map[string]*GpuMaterial{},
	}
}
