package cache

import (
	"github.com/galaco/kero/framework/graphics/adapter"
)

type GpuProp struct {
	Id       []*adapter.GpuMesh
	Material []GpuMaterial
}

func (prop *GpuProp) AddMesh(id *adapter.GpuMesh) {
	prop.Id = append(prop.Id, id)
}

func (prop *GpuProp) AddMaterial(mat GpuMaterial) {
	prop.Material = append(prop.Material, mat)
}

func NewGpuProp() *GpuProp {
	return &GpuProp{}
}
