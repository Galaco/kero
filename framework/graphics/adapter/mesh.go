package adapter

import (
	"github.com/galaco/gosigl"
	"github.com/go-gl/gl/v4.1-core/gl"
)

type Mesh interface {
	Vertices() []float32
	Normals() []float32
	UVs() []float32
	Tangents() []float32
	Indices() []uint
}

type GpuMesh *gosigl.VertexObject

func UploadMesh(mesh Mesh) GpuMesh {
	gpuResource := gosigl.NewMesh(mesh.Vertices())
	gosigl.CreateVertexAttribute(gpuResource, mesh.UVs(), 2)
	gosigl.CreateVertexAttribute(gpuResource, mesh.Normals(), 3)
	gosigl.CreateVertexAttribute(gpuResource, mesh.Tangents(), 4)
	gosigl.FinishMesh()

	return GpuMesh(gpuResource)
}

func UploadLightMesh(mesh Mesh) (GpuMesh, uint32) {
	gpuResource := gosigl.NewMesh(mesh.Vertices())

	var indexVbo uint32
	gl.GenBuffers(1, &indexVbo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, indexVbo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, 4 * len(mesh.Indices()), gl.Ptr(mesh.Indices()), gl.STATIC_DRAW)

	gosigl.FinishMesh()

	return GpuMesh(gpuResource), indexVbo
}

func DrawArray(offset int, num int) {
	gosigl.DrawArray(offset, num)
}

func DrawElements(num int) {
	gl.DrawElements(gl.TRIANGLES, int32(num), gl.UNSIGNED_INT, nil)
}

func DrawFace(offset int, num int) {
	DrawArray(offset, num)
}

func BindMesh(mesh *GpuMesh) {
	gosigl.BindMesh(*mesh)
}
