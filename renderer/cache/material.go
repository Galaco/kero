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
