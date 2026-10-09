package graphics

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"testing"
)

func TestNewError(t *testing.T) {
	tex := NewErrorTexture("error.vtf")

	if tex.Width() != 8 {
		t.Error("unexpected width")
	}
	if tex.Height() != 8 {
		t.Error("unexpected height")
	}

	expectedColourData := []uint8{
		255, 0, 255,
		255, 0, 255,
		255, 0, 255,
		255, 0, 255,
		0, 0, 0,
		0, 0, 0,
		0, 0, 0,
		0, 0, 0,

		255, 0, 255,
		255, 0, 255,
		255, 0, 255,
		255, 0, 255,
		0, 0, 0,
		0, 0, 0,
		0, 0, 0,
		0, 0, 0,

		255, 0, 255,
		255, 0, 255,
		255, 0, 255,
		255, 0, 255,
		0, 0, 0,
		0, 0, 0,
		0, 0, 0,
		0, 0, 0,

		255, 0, 255,
		255, 0, 255,
		255, 0, 255,
		255, 0, 255,
		0, 0, 0,
		0, 0, 0,
		0, 0, 0,
		0, 0, 0,

		0, 0, 0,
		0, 0, 0,
		0, 0, 0,
		0, 0, 0,
		255, 0, 255,
		255, 0, 255,
		255, 0, 255,
		255, 0, 255,

		0, 0, 0,
		0, 0, 0,
		0, 0, 0,
		0, 0, 0,
		255, 0, 255,
		255, 0, 255,
		255, 0, 255,
		255, 0, 255,

		0, 0, 0,
		0, 0, 0,
		0, 0, 0,
		0, 0, 0,
		255, 0, 255,
		255, 0, 255,
		255, 0, 255,
		255, 0, 255,

		0, 0, 0,
		0, 0, 0,
		0, 0, 0,
		0, 0, 0,
		255, 0, 255,
		255, 0, 255,
		255, 0, 255,
		255, 0, 255,
	}

	for idx, v := range expectedColourData {
		if tex.Image()[idx] != v {
			t.Error("unexpected colour data for error texture")
		}
	}
}

func TestColour2D_Format(t *testing.T) {
	tex := NewErrorTexture("error.vtf")
	if tex.Format() != gl.RGB {
		t.Error("unexpected error colour data format")
	}
}

func TestColour2D_PixelDataForFrame(t *testing.T) {
	tex := NewErrorTexture("error.vtf")

	expectedColourData := []uint8{
		255, 0, 255,
		255, 0, 255,
		255, 0, 255,
		255, 0, 255,
		0, 0, 0,
		0, 0, 0,
		0, 0, 0,
		0, 0, 0,

		255, 0, 255,
		255, 0, 255,
		255, 0, 255,
		255, 0, 255,
		0, 0, 0,
		0, 0, 0,
		0, 0, 0,
		0, 0, 0,

		255, 0, 255,
		255, 0, 255,
		255, 0, 255,
		255, 0, 255,
		0, 0, 0,
		0, 0, 0,
		0, 0, 0,
		0, 0, 0,

		255, 0, 255,
		255, 0, 255,
		255, 0, 255,
		255, 0, 255,
		0, 0, 0,
		0, 0, 0,
		0, 0, 0,
		0, 0, 0,

		0, 0, 0,
		0, 0, 0,
		0, 0, 0,
		0, 0, 0,
		255, 0, 255,
		255, 0, 255,
		255, 0, 255,
		255, 0, 255,

		0, 0, 0,
		0, 0, 0,
		0, 0, 0,
		0, 0, 0,
		255, 0, 255,
		255, 0, 255,
		255, 0, 255,
		255, 0, 255,

		0, 0, 0,
		0, 0, 0,
		0, 0, 0,
		0, 0, 0,
		255, 0, 255,
		255, 0, 255,
		255, 0, 255,
		255, 0, 255,

		0, 0, 0,
		0, 0, 0,
		0, 0, 0,
		0, 0, 0,
		255, 0, 255,
		255, 0, 255,
		255, 0, 255,
		255, 0, 255,
	}

	for idx, v := range expectedColourData {
		if tex.Image()[idx] != v {
			t.Error("unexpected colour data for error texture")
		}
	}
}

func TestTextureAtlas_Pack(t *testing.T) {
	atlas := NewTextureAtlas(0, 0)
	// Boxes of many sizes, as lightmaps are, and empty ones, as unlit faces have. Each box is filled with its index.
	seed := uint32(1)
	random := func(n int) int {
		seed = seed*1664525 + 1013904223
		return int(seed>>16) % n
	}
	numBoxes := 500
	for i := 0; i < numBoxes; i++ {
		w, h := 1+random(32), 1+random(32)
		if i%7 == 0 {
			w, h = 0, 0
		}
		colour := make([]uint8, w*h*4)
		for p := 0; p < w*h; p++ {
			colour[p*4], colour[p*4+1], colour[p*4+2] = uint8(i), uint8(i>>8), 0
		}
		atlas.AddRaw(w, h, colour)
	}
	atlas.Pack()

	for i := 0; i < numBoxes; i++ {
		a := atlas.AtlasEntry(i)
		if a.W == 0 || a.H == 0 {
			continue
		}
		if int(a.X)+a.W > atlas.Width() || int(a.Y)+a.H > atlas.Height() {
			t.Errorf("box %d at %v,%v %dx%d is outside the %dx%d atlas", i, a.X, a.Y, a.W, a.H, atlas.Width(), atlas.Height())
		}
		for j := i + 1; j < numBoxes; j++ {
			b := atlas.AtlasEntry(j)
			if b.W == 0 || b.H == 0 {
				continue
			}
			if int(a.X) < int(b.X)+b.W && int(b.X) < int(a.X)+a.W && int(a.Y) < int(b.Y)+b.H && int(b.Y) < int(a.Y)+a.H {
				t.Errorf("box %d at %v,%v %dx%d overlaps box %d at %v,%v %dx%d", i, a.X, a.Y, a.W, a.H, j, b.X, b.Y, b.W, b.H)
			}
		}
		// Each box's colour is where its entry says
		for y := int(a.Y); y < int(a.Y)+a.H; y++ {
			for x := int(a.X); x < int(a.X)+a.W; x++ {
				p := (y*atlas.Width() + x) * 4
				if got := int(atlas.Image()[p]) | int(atlas.Image()[p+1])<<8; got != i {
					t.Fatalf("pixel %d,%d of box %d holds box %d", x, y, i, got)
				}
			}
		}
	}
}
