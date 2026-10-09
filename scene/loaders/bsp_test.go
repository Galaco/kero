package loader

import (
	"reflect"
	"testing"

	"github.com/galaco/bsp/lump/primitive/dispinfo"
	"github.com/galaco/bsp/lump/primitive/dispvert"
	"github.com/galaco/bsp/lump/primitive/face"
	"github.com/galaco/bsp/lump/primitive/plane"
	"github.com/galaco/bsp/lump/primitive/texinfo"
	"github.com/galaco/kero/framework/graphics/mesh"
	"github.com/go-gl/mathgl/mgl32"
)

func TestGenerateDisplacementFace(t *testing.T) {
	// A 128x64 face, raised 10 units by a power 2 displacement that starts at its second corner
	corners := []mgl32.Vec3{{0, 0, 0}, {0, 64, 0}, {128, 64, 0}, {128, 0, 0}}
	dispVerts := make([]dispvert.DispVert, 25)
	for i := range dispVerts {
		dispVerts[i] = dispvert.DispVert{Vec: mgl32.Vec3{0, 0, 1}, Dist: 10, Alpha: float32(i * 10)}
	}
	structure := bspstructs{
		planes:    []plane.Plane{{Normal: mgl32.Vec3{0, 0, 1}}},
		vertexes:  corners,
		surfEdges: []int32{0, 1, 2, 3},
		edges:     [][2]uint16{{0, 1}, {1, 2}, {2, 3}, {3, 0}},
		texInfos:  []texinfo.TexInfo{{}},
		dispInfos: []dispinfo.DispInfo{{StartPosition: mgl32.Vec3{0, 64, 0}, Power: 2}},
		dispVerts: dispVerts,
	}
	f := &face.Face{NumEdges: 4}
	m := mesh.NewMesh()

	dispFace := generateDisplacementFace(f, &structure, m)

	// A vertex for each point of the 5x5 grid, and two triangles for each of its squares
	onFace := dispFace.DisplacementVertices()
	if len(onFace) != 5*5 || len(m.Vertices()) != 5*5*3 || len(m.BlendWeights()) != 5*5 {
		t.Fatalf("got %d vertices on the face, %d in the mesh and %d blend weights, want %d", len(onFace), len(m.Vertices())/3, len(m.BlendWeights()), 5*5)
	}
	if dispFace.Offset() != 0 || dispFace.Length() != 4*4*6 || len(m.Indices()) != 4*4*6 {
		t.Fatalf("got the face's indices at %d, %d of them, and %d in the mesh, want %d at 0", dispFace.Offset(), dispFace.Length(), len(m.Indices()), 4*4*6)
	}

	// The grid runs from the start corner towards the previous corner, then towards the next corner
	start, previous, next := corners[1], corners[0], corners[2]
	for i, v := range onFace {
		position := mgl32.Vec3{m.Vertices()[i*3], m.Vertices()[i*3+1], m.Vertices()[i*3+2]}
		if !position.ApproxEqual(v.Base.Add(mgl32.Vec3{0, 0, 10})) {
			t.Errorf("vertex %d: displaced to %v from %v on the face, want 10 units above it", i, position, v.Base)
		}
		expected := start.Add(previous.Sub(start).Mul(v.Grid.X())).Add(next.Sub(start).Mul(v.Grid.Y()))
		if !v.Base.ApproxEqual(expected) {
			t.Errorf("vertex %d: at %v on the face, want %v for grid position %v", i, v.Base, expected, v.Grid)
		}
		// Vertices are in the order of the displacement's vertices
		if grid := (mgl32.Vec2{float32(i%5) / 4, float32(i/5) / 4}); v.Grid != grid {
			t.Errorf("vertex %d: at grid position %v, want %v", i, v.Grid, grid)
		}
		if weight := m.BlendWeights()[i]; weight != dispVerts[i].Alpha/255 {
			t.Errorf("vertex %d: got blend weight %f, want %f", i, weight, dispVerts[i].Alpha/255)
		}
	}
	if onFace[0].Grid != (mgl32.Vec2{0, 0}) || !onFace[0].Base.ApproxEqual(start) {
		t.Errorf("first vertex: got grid %v at %v, want the start corner", onFace[0].Grid, onFace[0].Base)
	}

	// The first square's triangles join its corners at grid positions 0,0 0,1 1,1 and 1,0
	if first := m.Indices()[:6]; !reflect.DeepEqual(first, []uint32{0, 5, 6, 0, 6, 1}) {
		t.Errorf("got the first square's triangles %v, want [0 5 6 0 6 1]", first)
	}

	// The face's bounds hold its displaced vertices
	if mins, maxs := dispFace.Bounds(); mins != (mgl32.Vec3{0, 0, 10}) || maxs != (mgl32.Vec3{128, 64, 10}) {
		t.Errorf("got bounds %v to %v, want 0,0,10 to 128,64,10", mins, maxs)
	}
}

func TestLuxelToLightmap(t *testing.T) {
	cases := []struct {
		colour   uint8
		exponent int8
		expected uint8
	}{
		{0, 0, 0},
		// Full light is stored at half brightness, for shaders to double
		{255, 0, 128},
		{128, 1, 128},
		// Dim light is gamma corrected
		{64, -2, 36},
		{255, -8, 10},
		// Light is stored up to 4 times full brightness
		{255, 2, 239},
		{200, 5, 239},
	}
	for _, c := range cases {
		if actual := luxelToLightmap(c.colour, c.exponent); actual != c.expected {
			t.Errorf("luxel %d with exponent %d: got %d, want %d", c.colour, c.exponent, actual, c.expected)
		}
	}
}
