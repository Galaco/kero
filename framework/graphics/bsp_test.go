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
	uvs := DisplacementLightmapCoords(vertices, f, &texinfo.TexInfo{}, 100, 50, 10, 20)
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

	// SURFDRAW_NOLIGHT
	uvs = DisplacementLightmapCoords(vertices, f, &texinfo.TexInfo{Flags: 1}, 100, 50, 10, 20)
	for i, uv := range uvs {
		if uv != 0.5 {
			t.Errorf("unlit lightmap coordinate %d: got %f, want 0.5", i, uv)
		}
	}
}
