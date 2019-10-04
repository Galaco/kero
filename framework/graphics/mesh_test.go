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
