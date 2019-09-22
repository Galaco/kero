package graphics

import (
	"github.com/galaco/kero/framework/graphics/util"
)

// Mesh
type Mesh struct {
	vertices []float32
	normals  []float32
	uvs      []float32
	tangents []float32
}

// AddVertex
func (mesh *Mesh) AddVertex(vertex ...float32) {
	mesh.vertices = append(mesh.vertices, vertex...)
}

// AddNormal
func (mesh *Mesh) AddNormal(normal ...float32) {
	mesh.normals = append(mesh.normals, normal...)
}

// AddUV
func (mesh *Mesh) AddUV(uv ...float32) {
	mesh.uvs = append(mesh.uvs, uv...)
}

// AddTangent
func (mesh *Mesh) AddTangent(tangent ...float32) {
	mesh.tangents = append(mesh.tangents, tangent...)
}

// Vertices
func (mesh *Mesh) Vertices() []float32 {
	return mesh.vertices
}

// Normals
func (mesh *Mesh) Normals() []float32 {
	return mesh.normals
}

// UVs
func (mesh *Mesh) UVs() []float32 {
	return mesh.uvs
}

// Tangents
func (mesh *Mesh) Tangents() []float32 {
	return mesh.tangents
}

// GenerateTangents
func (mesh *Mesh) GenerateTangents() {
	mesh.tangents = util.GenerateTangents(mesh.Vertices(), mesh.Normals(), mesh.UVs())
}

// NewMesh
func NewMesh() *Mesh {
	return &Mesh{}
}