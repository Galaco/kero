package graphics

import (
	"github.com/galaco/vtf/format"
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
		if tex.PixelDataForFrame(0)[idx] != v {
			t.Error("unexpected colour data for error texture")
		}
	}
}

func TestColour2D_Format(t *testing.T) {
	tex := NewErrorTexture("error.vtf")
	if tex.Format() != uint32(format.RGB888) {
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
