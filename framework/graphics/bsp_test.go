package graphics

import (
	"testing"

	"github.com/galaco/bsp/lump/primitive/face"
	"github.com/galaco/bsp/lump/primitive/texinfo"
	"github.com/go-gl/mathgl/mgl32"
)

func TestDisplacementLightmapCoords(t *testing.T) {
	vertices := []DisplacementVertex{
		{Grid: mgl32.Vec2{0, 0}},
		{Grid: mgl32.Vec2{1, 0}},
		{Grid: mgl32.Vec2{0, 1}},
		{Grid: mgl32.Vec2{1, 1}},
		{Grid: mgl32.Vec2{0.5, 0.25}},
	}
	f := &face.Face{LightmapTextureSizeInLuxels: [2]int32{8, 4}}

	// The face's lightmap is 9x5 luxels, at 10,20 in a 100x50 atlas. The grid's corners are the corner luxels' centres.
	uvs := DisplacementLightmapCoords(vertices, f, 100, 50, 10, 20)
	expected := []float32{
		10.5 / 100, 20.5 / 50,
		18.5 / 100, 20.5 / 50,
		10.5 / 100, 24.5 / 50,
		18.5 / 100, 24.5 / 50,
		14.5 / 100, 21.5 / 50,
	}
	for i := range expected {
		if mgl32.Abs(uvs[i]-expected[i]) > 1e-6 {
			t.Errorf("lightmap coordinate %d: got %f, want %f", i, uvs[i], expected[i])
		}
	}

	// A displacement without samples is lit by its single luxel
	f.Lightofs = -1
	uvs = DisplacementLightmapCoords(vertices, f, 100, 50, 10, 20)
	for i := range vertices {
		if mgl32.Abs(uvs[i*2]-10.5/100) > 1e-6 || mgl32.Abs(uvs[i*2+1]-20.5/50) > 1e-6 {
			t.Errorf("unlit lightmap coordinate %d: got %f,%f, want its luxel's centre", i, uvs[i*2], uvs[i*2+1])
		}
	}
}

func TestLightmapCoordsForFaceFromTexInfo(t *testing.T) {
	// A face lit one luxel per unit along x and y. Its lightmap is 3x2 luxels, at 10,20 in a 100x50 atlas.
	vertices := []float32{0, 0, 0, 2, 1, 0}
	f := &face.Face{LightmapTextureSizeInLuxels: [2]int32{2, 1}}
	// SURF_LIGHT marks a face that lights the map; it is lit like any other
	tx := &texinfo.TexInfo{Flags: 0x1, LightmapVecsLuxelsPerWorldUnits: [2][4]float32{{1, 0, 0, 0}, {0, 1, 0, 0}}}
	expected := []float32{10.5 / 100, 20.5 / 50, 12.5 / 100, 21.5 / 50}

	uvs := LightmapCoordsForFaceFromTexInfo(vertices, f, tx, 100, 50, 10, 20)
	for i := range expected {
		if mgl32.Abs(uvs[i]-expected[i]) > 1e-6 {
			t.Errorf("lightmap coordinate %d: got %f, want %f", i, uvs[i], expected[i])
		}
	}

	// A face without samples is lit by its single luxel
	f.Lightofs = -1
	uvs = LightmapCoordsForFaceFromTexInfo(vertices, f, tx, 100, 50, 10, 20)
	for i := range vertices[:2] {
		if mgl32.Abs(uvs[i*2]-10.5/100) > 1e-6 || mgl32.Abs(uvs[i*2+1]-20.5/50) > 1e-6 {
			t.Errorf("unlit lightmap coordinate %d: got %f,%f, want its luxel's centre", i, uvs[i*2], uvs[i*2+1])
		}
	}
}
