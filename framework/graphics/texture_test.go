package graphics

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"

	"github.com/galaco/vtf/v2"
	"github.com/galaco/vtf/v2/format"
	"github.com/go-gl/gl/v4.1-core/gl"
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

func TestTextureAtlas_Pages(t *testing.T) {
	atlas := NewPagedTextureAtlas(3)
	// Each box's image for a page is filled with its index and the page
	sizes := [][2]int{{4, 3}, {2, 5}, {0, 0}, {3, 3}}
	for i, size := range sizes {
		w, h := size[0], size[1]
		colour := make([]uint8, 0, w*h*4*3)
		for page := 0; page < 3; page++ {
			for p := 0; p < w*h; p++ {
				colour = append(colour, uint8(i), uint8(page), 0, 255)
			}
		}
		atlas.AddRaw(w, h, colour)
	}
	atlas.Pack()

	if atlas.Height() != atlas.PageHeight()*3 {
		t.Fatalf("got a %d high atlas of %d high pages, want 3 pages", atlas.Height(), atlas.PageHeight())
	}
	for i, size := range sizes {
		a := atlas.AtlasEntry(i)
		if a.W != size[0] || a.H != size[1] {
			t.Fatalf("box %d: got %dx%d, want %dx%d", i, a.W, a.H, size[0], size[1])
		}
		if int(a.Y)+a.H > atlas.PageHeight() {
			t.Errorf("box %d at %v,%v %dx%d is outside the first %d high page", i, a.X, a.Y, a.W, a.H, atlas.PageHeight())
		}
		// Each page holds the box's image for that page, at the same place
		for page := 0; page < 3; page++ {
			for y := int(a.Y); y < int(a.Y)+a.H; y++ {
				for x := int(a.X); x < int(a.X)+a.W; x++ {
					p := ((y+page*atlas.PageHeight())*atlas.Width() + x) * 4
					if box, got := int(atlas.Image()[p]), int(atlas.Image()[p+1]); box != i || got != page {
						t.Fatalf("pixel %d,%d of box %d in page %d holds box %d's image for page %d", x, y, i, page, box, got)
					}
				}
			}
		}
	}
}

// textureFiles is a filesystem of textures by path
type textureFiles map[string][]byte

func (files textureFiles) GetFile(path string) (io.Reader, error) {
	data, ok := files[path]
	if !ok {
		return nil, errors.New("not found: " + path)
	}
	return bytes.NewReader(data), nil
}

func TestLoadTexture_Mipmaps(t *testing.T) {
	// A 4x2 BGR888 texture with every mipmap. Each mipmap is filled with its level, and they are stored smallest first.
	header := vtf.Header{}
	header.Signature = [4]byte{'V', 'T', 'F', 0}
	header.Version = [2]uint32{7, 2}
	header.HeaderSize = 80
	header.Width, header.Height = 4, 2
	header.Frames = 1
	header.HighResImageFormat = format.BGR888
	header.MipmapCount = 3
	header.LowResImageFormat = format.None
	header.Depth = 1
	file := bytes.Buffer{}
	if err := binary.Write(&file, binary.LittleEndian, header); err != nil {
		t.Fatal(err)
	}
	file.Write(make([]byte, int(header.HeaderSize)-file.Len()))
	sizes := []int{4 * 2 * 3, 2 * 1 * 3, 1 * 1 * 3}
	for level := len(sizes) - 1; level >= 0; level-- {
		file.Write(bytes.Repeat([]byte{byte(level)}, sizes[level]))
	}

	texture, err := LoadTexture(textureFiles{"materials/foo.vtf": file.Bytes()}, "foo")
	if err != nil {
		t.Fatal(err)
	}

	if texture.Width() != 4 || texture.Height() != 2 || !bytes.Equal(texture.Image(), bytes.Repeat([]byte{0}, sizes[0])) {
		t.Errorf("got a %dx%d texture of %v, want the largest mipmap", texture.Width(), texture.Height(), texture.Image())
	}
	mipmaps := texture.Mipmaps()
	if len(mipmaps) != len(sizes) {
		t.Fatalf("got %d mipmaps, want %d", len(mipmaps), len(sizes))
	}
	// From the largest
	for level, mipmap := range mipmaps {
		if expected := bytes.Repeat([]byte{byte(level)}, sizes[level]); !bytes.Equal(mipmap, expected) {
			t.Errorf("mipmap %d: got %v, want %v", level, mipmap, expected)
		}
	}
}
