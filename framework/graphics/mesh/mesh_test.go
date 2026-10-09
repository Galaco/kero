package mesh

import (
	"math"
	"reflect"
	"testing"

	"github.com/go-gl/mathgl/mgl32"
)

func TestNewMesh(t *testing.T) {
	if reflect.TypeOf(NewMesh()) != reflect.TypeOf(&BasicMesh{}) {
		t.Errorf("unexpected type returned for NewMesh. Expected: %s, but received: %s", reflect.TypeOf(&BasicMesh{}), reflect.TypeOf(NewMesh()))
	}
}

func TestMesh_AddNormal(t *testing.T) {
	sut := BasicMesh{}
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
	sut := BasicMesh{}
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
	sut := BasicMesh{}
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
	sut := BasicMesh{}
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
	sut := BasicMesh{}
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
	sut := BasicMesh{}
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

func TestMesh_GenerateTangents(t *testing.T) {
	// A unit quad in the xy plane facing +z, with u along +x and v along +y
	corners := [][3]float32{{0, 0, 0}, {1, 0, 0}, {1, 1, 0}, {0, 1, 0}}
	cornerUVs := [][2]float32{{0, 0}, {1, 0}, {1, 1}, {0, 1}}
	quad := func(order []int, indices []uint32, flipU bool) *BasicMesh {
		sut := NewMesh()
		for _, c := range order {
			sut.AddVertex(corners[c][:]...)
			sut.AddNormal(0, 0, 1)
			u := cornerUVs[c][0]
			if flipU {
				u = 1 - u
			}
			sut.AddUV(u, cornerUVs[c][1])
		}
		sut.AddIndice(indices...)
		return sut
	}

	cases := []struct {
		name     string
		mesh     *BasicMesh
		expected [4]float32
	}{
		{"unindexed", quad([]int{0, 1, 2, 0, 2, 3}, nil, false), [4]float32{1, 0, 0, 1}},
		{"indexed", quad([]int{0, 1, 2, 3}, []uint32{0, 1, 2, 0, 2, 3}, false), [4]float32{1, 0, 0, 1}},
		{"mirrored", quad([]int{0, 1, 2, 3}, []uint32{0, 1, 2, 0, 2, 3}, true), [4]float32{-1, 0, 0, -1}},
	}
	for _, c := range cases {
		c.mesh.GenerateTangents()
		tangents := c.mesh.Tangents()
		numVertices := len(c.mesh.Vertices()) / 3
		if len(tangents) != numVertices*4 {
			t.Fatalf("%s: expected %d tangent floats, got %d", c.name, numVertices*4, len(tangents))
		}
		for v := 0; v < numVertices; v++ {
			for i := 0; i < 4; i++ {
				if math.Abs(float64(tangents[v*4+i]-c.expected[i])) > 1e-5 {
					t.Errorf("%s: vertex %d: expected tangent %v, got %v", c.name, v, c.expected, tangents[v*4:v*4+4])
					break
				}
			}
		}
	}
}

func TestMesh_GenerateTangentsDegenerate(t *testing.T) {
	// Fewer vertices than a triangle
	sut := NewMesh()
	sut.AddVertex(0, 0, 0, 1, 0, 0)
	sut.AddNormal(0, 0, 1, 0, 0, 1)
	sut.AddUV(0, 0, 1, 0)
	sut.GenerateTangents()
	if len(sut.Tangents()) != 8 {
		t.Fatalf("expected 8 tangent floats, got %d", len(sut.Tangents()))
	}

	// A triangle with no texture area still gets unit tangents perpendicular to its normal
	sut = NewMesh()
	sut.AddVertex(0, 0, 0, 1, 0, 0, 0, 1, 0)
	sut.AddNormal(0, 0, 1, 0, 0, 1, 0, 0, 1)
	sut.AddUV(0, 0, 0, 0, 0, 0)
	sut.GenerateTangents()
	tangents := sut.Tangents()
	for v := 0; v < 3; v++ {
		tangent := mgl32.Vec3{tangents[v*4], tangents[v*4+1], tangents[v*4+2]}
		if math.Abs(float64(tangent.Len()-1)) > 1e-5 || math.Abs(float64(tangent.Z())) > 1e-5 {
			t.Errorf("vertex %d: expected a unit tangent in the xy plane, got %v", v, tangent)
		}
	}
}
