package graphics

import (
	"github.com/galaco/kero/framework/graphics/util"
)

type Mesh interface {
	Vertices() []float32
	Normals() []float32
	UVs() []float32
	Tangents() []float32
	Indices() []uint32
}

// BasicMesh
type BasicMesh struct {
	vertices    []float32
	normals     []float32
	uvs         []float32
	lightmapUVs []float32
	tangents    []float32
	indices     []uint32
}

// AddVertex
func (mesh *BasicMesh) AddVertex(vertex ...float32) {
	mesh.vertices = append(mesh.vertices, vertex...)
}

// AddNormal
func (mesh *BasicMesh) AddNormal(normal ...float32) {
	mesh.normals = append(mesh.normals, normal...)
}

// AddUV
func (mesh *BasicMesh) AddUV(uv ...float32) {
	mesh.uvs = append(mesh.uvs, uv...)
}

// AddLightmapUV
func (mesh *BasicMesh) AddLightmapUV(uv ...float32) {
	mesh.lightmapUVs = append(mesh.lightmapUVs, uv...)
}

// AddTangent
func (mesh *BasicMesh) AddTangent(tangent ...float32) {
	mesh.tangents = append(mesh.tangents, tangent...)
}

// AddIndice
func (mesh *BasicMesh) AddIndice(indice ...uint32) {
	mesh.indices = append(mesh.indices, indice...)
}

// Vertices
func (mesh *BasicMesh) Vertices() []float32 {
	return mesh.vertices
}

// Normals
func (mesh *BasicMesh) Normals() []float32 {
	return mesh.normals
}

// UVs
func (mesh *BasicMesh) UVs() []float32 {
	return mesh.uvs
}

// LightmapUVs
func (mesh *BasicMesh) LightmapUVs() []float32 {
	return mesh.lightmapUVs
}

// Tangents
func (mesh *BasicMesh) Tangents() []float32 {
	return mesh.tangents
}

// Indices
func (mesh *BasicMesh) Indices() []uint32 {
	return mesh.indices
}

// GenerateTangents
func (mesh *BasicMesh) GenerateTangents() {
	mesh.tangents = util.GenerateTangents(mesh.Vertices(), mesh.Normals(), mesh.UVs())
}

// NewMesh
func NewMesh() *BasicMesh {
	return &BasicMesh{}
}
