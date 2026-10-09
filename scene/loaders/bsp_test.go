package loader

import (
	"math"
	"reflect"
	"testing"

	"github.com/galaco/bsp/lump/primitive/common"
	"github.com/galaco/bsp/lump/primitive/dispinfo"
	"github.com/galaco/bsp/lump/primitive/dispvert"
	"github.com/galaco/bsp/lump/primitive/face"
	"github.com/galaco/bsp/lump/primitive/plane"
	"github.com/galaco/bsp/lump/primitive/texinfo"
	"github.com/galaco/kero/framework/graphics"
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

func TestBumpedLuxelsToLightmap(t *testing.T) {
	full := common.ColorRGBExponent32{R: 255, G: 255, B: 255, Exponent: 0}
	twice := common.ColorRGBExponent32{R: 255, G: 255, B: 255, Exponent: 1}
	none := common.ColorRGBExponent32{}

	// Light equal from every direction is stored as the face's lightmap is
	lightmaps := bumpedLuxelsToLightmap(full, [3]common.ColorRGBExponent32{full, full, full})
	for direction, lightmap := range lightmaps {
		if lightmap != [3]uint8{128, 128, 128} {
			t.Errorf("even light from direction %d: got %v, want the face's lightmap, 128", direction, lightmap)
		}
	}

	// Light from each direction is scaled so that its average is the face's lightmap
	lightmaps = bumpedLuxelsToLightmap(full, [3]common.ColorRGBExponent32{twice, full, none})
	if expected := [3][3]uint8{{255, 255, 255}, {128, 128, 128}, {0, 0, 0}}; lightmaps != expected {
		t.Errorf("light from two directions: got %v, want %v", lightmaps, expected)
	}

	// An unlit face is unlit from every direction
	if lightmaps = bumpedLuxelsToLightmap(none, [3]common.ColorRGBExponent32{none, none, none}); lightmaps != [3][3]uint8{} {
		t.Errorf("unlit: got %v, want nothing", lightmaps)
	}

	// The average of the directions' lightmaps is the face's lightmap, unless one is clamped at full brightness
	seed := uint32(1)
	random := func(n int) int {
		seed = seed*1664525 + 1013904223
		return int(seed>>16) % n
	}
	luxel := func() common.ColorRGBExponent32 {
		return common.ColorRGBExponent32{R: uint8(random(256)), G: uint8(random(256)), B: uint8(random(256)), Exponent: int8(random(4) - 3)}
	}
	for i := 0; i < 1000; i++ {
		flat := luxel()
		lightmaps := bumpedLuxelsToLightmap(flat, [3]common.ColorRGBExponent32{luxel(), luxel(), luxel()})
		flatChannels := [3]uint8{flat.R, flat.G, flat.B}
		for channel := 0; channel < 3; channel++ {
			if lightmaps[0][channel] == 255 || lightmaps[1][channel] == 255 || lightmaps[2][channel] == 255 {
				continue
			}
			average := (float64(lightmaps[0][channel]) + float64(lightmaps[1][channel]) + float64(lightmaps[2][channel])) / 3
			if lightmaps[0][channel]+lightmaps[1][channel]+lightmaps[2][channel] == 0 {
				continue
			}
			if expected := float64(luxelToLightmap(flatChannels[channel], flat.Exponent)); math.Abs(average-expected) > 1 {
				t.Fatalf("luxel %v: directions average %f, want the face's lightmap, %f", flat, average, expected)
			}
		}
	}
}

func TestLightmapFromFace(t *testing.T) {
	// A 2x1 luxel face
	f := &face.Face{Lightofs: 4 * 4, LightmapTextureSizeInLuxels: [2]int32{1, 0}}
	full := common.ColorRGBExponent32{R: 255, G: 255, B: 255}
	half := common.ColorRGBExponent32{R: 255, G: 255, B: 255, Exponent: -1}
	twice := common.ColorRGBExponent32{R: 255, G: 255, B: 255, Exponent: 1}
	none := common.ColorRGBExponent32{}
	// Four samples before the face's, then its lightmap and the lightmap of each bump basis direction
	samples := []common.ColorRGBExponent32{none, none, none, none, full, half, twice, full, full, half, none, half}

	// The face's lightmap is in every page
	width, height, colour := lightmapFromFace(f, &texinfo.TexInfo{}, samples)
	if width != 2 || height != 1 || len(colour) != 2*4*graphics.LightmapPages {
		t.Fatalf("got a %dx%d lightmap of %d bytes, want 2x1 in %d pages", width, height, len(colour), graphics.LightmapPages)
	}
	flatPage := []uint8{128, 128, 128, 255, luxelToLightmap(255, -1), luxelToLightmap(255, -1), luxelToLightmap(255, -1), 255}
	for page := 0; page < graphics.LightmapPages; page++ {
		if got := colour[page*8 : page*8+8]; !reflect.DeepEqual(got, flatPage) {
			t.Errorf("page %d of a face not lit for bump mapping: got %v, want %v", page, got, flatPage)
		}
	}

	// A face lit for bump mapping has the lightmap of each direction in the pages after its own
	_, _, colour = lightmapFromFace(f, &texinfo.TexInfo{Flags: surfBumpLight}, samples)
	if got := colour[:8]; !reflect.DeepEqual(got, flatPage) {
		t.Errorf("first page of a face lit for bump mapping: got %v, want its lightmap %v", got, flatPage)
	}
	for luxel := 0; luxel < 2; luxel++ {
		expected := bumpedLuxelsToLightmap(samples[4+luxel], [3]common.ColorRGBExponent32{samples[6+luxel], samples[8+luxel], samples[10+luxel]})
		for direction := 0; direction < 3; direction++ {
			offset := (direction+1)*8 + luxel*4
			if got := [3]uint8{colour[offset], colour[offset+1], colour[offset+2]}; got != expected[direction] {
				t.Errorf("luxel %d lit from direction %d: got %v, want %v", luxel, direction, got, expected[direction])
			}
		}
	}

	// A face without samples has a single luxel in every page: fully lit if it isn't meant to be lit, and unlit if no
	// light reached it
	for _, c := range []struct {
		name  string
		flags int32
		luxel uint8
	}{{"a face that isn't lit", surfNoLight, 128}, {"a face no light reached", 0, 0}} {
		width, height, colour = lightmapFromFace(&face.Face{Lightofs: -1}, &texinfo.TexInfo{Flags: c.flags}, samples)
		if width != 1 || height != 1 || len(colour) != 4*graphics.LightmapPages {
			t.Fatalf("%s: got a %dx%d lightmap of %d bytes, want 1x1 in %d pages", c.name, width, height, len(colour), graphics.LightmapPages)
		}
		for page := 0; page < graphics.LightmapPages; page++ {
			if got, want := colour[page*4:page*4+4], []uint8{c.luxel, c.luxel, c.luxel, 255}; !reflect.DeepEqual(got, want) {
				t.Errorf("%s: page %d is %v, want %v", c.name, page, got, want)
			}
		}
	}
}
