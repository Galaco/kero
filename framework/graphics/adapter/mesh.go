package adapter

import "github.com/galaco/gosigl"

type Mesh interface {
	Vertices() []float32
	Normals() []float32
	UVs() []float32
	Tangents() []float32
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

func DrawArray(offset int, num int) {
	gosigl.DrawArray(offset, num)
}

func DrawFace(offset int, num int, textureId uint32) {
	BindTexture(textureId)
	DrawArray(offset, num)
}

func BindMesh(mesh *GpuMesh) {
	gosigl.BindMesh(*mesh)
}
