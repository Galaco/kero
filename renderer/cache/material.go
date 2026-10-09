package cache

import (
	"github.com/galaco/kero/framework/graphics"
)

const (
	ErrorMaterialPath = "materials/error.vmt"
)

type GpuMaterial struct {
	Diffuse    uint32
	Diffuse2   uint32 // Second texture for blend materials (WorldVertexTransition)
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
