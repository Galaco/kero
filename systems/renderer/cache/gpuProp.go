package cache

import "github.com/galaco/kero/framework/graphics"

type GpuProp struct {
	Id       []*graphics.GpuMesh
	Material []GpuMaterial
}

func (prop *GpuProp) AddMesh(id *graphics.GpuMesh) {
	prop.Id = append(prop.Id, id)
}

func (prop *GpuProp) AddMaterial(mat GpuMaterial) {
	prop.Material = append(prop.Material, mat)
}

func NewGpuProp() *GpuProp {
	return &GpuProp{}
}
