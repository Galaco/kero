package graphics

import (
	"reflect"
	"testing"
)

func TestNewMesh(t *testing.T) {
	if reflect.TypeOf(NewMesh()) != reflect.TypeOf(&Mesh{}) {
		t.Errorf("unexpected type returned for NewMesh. Expected: %s, but received: %s", reflect.TypeOf(&Mesh{}), reflect.TypeOf(NewMesh()))
	}
}

func TestMesh_AddNormal(t *testing.T) {
	sut := Mesh{}
	expected := []float32{
		1, 2, 3, 4,
	}
	sut.AddNormal(expected...)

	for i := 0; i < len(expected); i++ {
		if sut.Normals()[i] != expected[i] {
			t.Error("unexpected normal")
		}
	}
}

func TestMesh_AddTextureCoordinate(t *testing.T) {
	sut := Mesh{}
	expected := []float32{
		1, 2, 3, 4,
	}
	sut.AddUV(expected...)

	for i := 0; i < len(expected); i++ {
		if sut.UVs()[i] != expected[i] {
			t.Error("unexpected texture coordinate")
		}
	}
}

func TestMesh_AddVertex(t *testing.T) {
	sut := Mesh{}
	expected := []float32{
		1, 2, 3, 4,
	}
	sut.AddVertex(expected...)

	for i := 0; i < len(expected); i++ {
		if sut.Vertices()[i] != expected[i] {
			t.Error("unexpected vertex")
		}
	}
}

func TestMesh_Normals(t *testing.T) {
	sut := Mesh{}
	expected := []float32{
		1, 2, 3, 4,
	}
	sut.AddNormal(expected...)

	for i := 0; i < len(expected); i++ {
		if sut.Normals()[i] != expected[i] {
			t.Error("unexpected normal")
		}
	}
}

func TestMesh_TextureCoordinates(t *testing.T) {
	sut := Mesh{}
	expected := []float32{
		1, 2, 3, 4,
	}
	sut.AddUV(expected...)

	for i := 0; i < len(expected); i++ {
		if sut.UVs()[i] != expected[i] {
			t.Error("unexpected texture coordinate")
		}
	}
}

func TestMesh_Vertices(t *testing.T) {
	sut := Mesh{}
	expected := []float32{
		1, 2, 3, 4,
	}
	sut.AddVertex(expected...)

	for i := 0; i < len(expected); i++ {
		if sut.Vertices()[i] != expected[i] {
			t.Error("unexpected vertex")
		}
	}
}

func TestMeshFace_Length(t *testing.T) {
	sut := NewMeshFace(32, 64)

	if sut.Length() != 64 {
		t.Error("unexpected length for face")
	}
}

func TestMeshFace_Offset(t *testing.T) {
	sut := NewMeshFace(32, 64)

	if sut.Offset() != 32 {
		t.Error("unexpected offset for face")
	}
}

func TestNewMeshFace(t *testing.T) {
	sut := NewMeshFace(32, 64)
	if reflect.TypeOf(sut) != reflect.TypeOf(MeshFace{}) {
		t.Errorf("unexpceted type returned. Expected %s, but received: %s", reflect.TypeOf(MeshFace{}), reflect.TypeOf(sut))
	}
}
