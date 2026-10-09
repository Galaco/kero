package cache

import (
	"github.com/galaco/kero/framework/graphics/adapter"
)

// GpuProp is a model uploaded to the GPU. Its sub-meshes share Mesh, and are drawn with Material.
type GpuProp struct {
	Mesh     adapter.GpuMesh
	Material []GpuMaterial
}

func (prop *GpuProp) AddMaterial(mat GpuMaterial) {
	prop.Material = append(prop.Material, mat)
}
