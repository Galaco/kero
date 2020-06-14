package graphics

import (
	"github.com/galaco/vtf"
	"github.com/galaco/vtf/format"
	"log"
	"math"
	"sort"
	"strings"
)

// Texture2D is a material defined by raw/computed colour data
type Texture2D struct {
	filePath string
	width    int
	height   int
	format   uint32
	colour   []uint8
}

// Format returns colour format
func (texture *Texture2D) Format() uint32 {
	return texture.format
}

// Width
func (texture *Texture2D) Width() int {
	return texture.width
}

// Height
func (texture *Texture2D) Height() int {
	return texture.height
}

// Image returns raw colour data
func (texture *Texture2D) Image() []uint8 {
	return texture.colour
}

// LoadTexture
func LoadTexture(fs VirtualFileSystem, filePath string) (*Texture2D, error) {
	if !strings.HasSuffix(filePath, ExtensionVtf) {
		filePath = filePath + ExtensionVtf
	}
	return readVtf(fs, BasePathMaterial+filePath)
}

func NewTexture(filePath string, width, height int, format uint32, colour []uint8) *Texture2D {
	return &Texture2D{
		filePath: filePath,
		width:    width,
		height:   height,
		format:   textureFormatFromVtfFormat(format),
		colour:   colour,
	}
}

// NewError returns new Error material
func NewErrorTexture(name string) *Texture2D {
	return NewTexture(
		name,
		8,
		8,
		uint32(format.RGB888),
		[]uint8{
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
		})
}

// readVtf
func readVtf(fs VirtualFileSystem, path string) (*Texture2D, error) {
	stream, err := fs.GetFile(path)
	if err != nil {
		return nil, err
	}

	// Attempt to parse the vtf into color data we can use,
	// if this fails (it shouldn't) we can treat it like it was missing
	read, err := vtf.ReadFromStream(stream)
	if err != nil {
		return nil, err
	}

	return NewTexture(path,
			int(read.Header().Width),
			int(read.Header().Height),
			read.Header().HighResImageFormat,
			read.HighestResolutionImageForFrame(0)),
		nil
}

// TextureAtlas is a simple 2d texture atlas.
// Does NOT support transparency
type TextureAtlas struct {
	rectangles []AtlasTexture

	width, height int
	colour        []uint8
	format        uint32
}

func (atlas *TextureAtlas) AddRaw(width, height int, colour []uint8) *AtlasTexture {
	atlas.rectangles = append(atlas.rectangles, AtlasTexture{
		W:      width,
		H:      height,
		X:      0,
		Y:      0,
		colour: colour,
	})

	return &(atlas.rectangles[len(atlas.rectangles)-1])
}

func (atlas *TextureAtlas) Pack() {
	// calculate total box area and maximum box width
	area := 0
	maxWidth := 0

	for _, box := range atlas.rectangles {
		area += box.W * box.H
		maxWidth = int(math.Max(float64(maxWidth), float64(box.W)))
	}

	// sort the boxes for insertion by height, descending
	sort.Slice(atlas.rectangles, func(i, j int) bool {
		return atlas.rectangles[i].H > atlas.rectangles[j].H
	})

	// aim for a squarish resulting container,
	// slightly adjusted for sub-100% space utilization
	startWidth := math.Max(math.Ceil(math.Sqrt(float64(area)/0.95)), float64(maxWidth))

	// start with a single empty space, unbounded at the bottom
	spaces := []atlasSpace{
		{x: 0, y: 0, w: int(startWidth), h: 99999999},
	}
	packed := make([]AtlasTexture, 0)

	for _, box := range atlas.rectangles {
		// look through spaces backwards so that we check smaller spaces first
		for i := len(spaces) - 1; i >= 0; i-- {
			space := spaces[i]

			// look for empty spaces that can accommodate the current box
			if box.W > space.w || box.H > space.h {
				continue
			}

			// found the space; add the box to its top-left corner
			// |-------|-------|
			// |  box  |       |
			// |_______|       |
			// |         space |
			// |_______________|
			packed = append(packed, AtlasTexture{W: box.W, H: box.H, X: float32(space.x), Y: float32(space.y)})

			// Insert colour data here to skip some duplication

			if int(box.W) == space.w && int(box.H) == space.h {
				// space matches the box exactly; remove it
				last := spaces[len(spaces)-1]
				spaces = spaces[:len(spaces)-1]

				if i < len(spaces) {
					spaces[i] = last
				}
			} else if box.H == space.h {
				// space matches the box height; update it accordingly
				// |-------|---------------|
				// |  box  | updated space |
				// |_______|_______________|
				space.x += box.W
				space.w -= box.W
			} else if box.W == space.w {
				// space matches the box width; update it accordingly
				// |---------------|
				// |      box      |
				// |_______________|
				// | updated space |
				// |_______________|
				space.y += box.H
				space.h -= box.H
			} else {
				// otherwise the box splits the space into two spaces
				// |-------|-----------|
				// |  box  | new space |
				// |_______|___________|
				// | updated space     |
				// |___________________|
				spaces = append(spaces, atlasSpace{
					x: space.x + box.W,
					y: space.y,
					w: space.w - box.W,
					h: box.H,
				})
				space.y += box.H
				space.h -= box.H
			}
			break
		}
	}
	log.Println(packed)
}

func NewTextureAtlas(width, height int) *TextureAtlas {
	return &TextureAtlas{
		width:      width,
		height:     height,
		rectangles: []AtlasTexture{},
		colour:     make([]uint8, width*height*3),
		format:     uint32(format.RGB888),
	}
}

type AtlasTexture struct {
	W, H int
	X, Y float32

	colour []uint8
}

type atlasSpace struct {
	x, y, w, h int
}
