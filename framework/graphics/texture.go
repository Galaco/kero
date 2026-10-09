package graphics

import (
	"github.com/galaco/kero/framework/graphics/adapter"
	"github.com/galaco/vtf/v2"
	"github.com/galaco/vtf/v2/format"
	"math"
	"sort"
	"strings"
)

type Texture adapter.Texture

// Texture2D is a material defined by raw/computed colour data
type Texture2D struct {
	filePath string
	width    int
	height   int
	format   uint32
	colour   []uint8
	// mipmaps are smaller copies of colour, each half the size of the one before. It is empty if there are none.
	mipmaps [][]uint8
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

// Mipmaps returns the texture's colour data, then each of its smaller copies, from the largest
func (texture *Texture2D) Mipmaps() [][]uint8 {
	return append([][]uint8{texture.colour}, texture.mipmaps...)
}

// Free color data from memory
func (texture *Texture2D) Release() {
	texture.colour = nil
	texture.mipmaps = nil
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
		format:   adapter.TextureFormatFromVtfFormat(format),
		colour:   colour,
	}
}

// NewError returns new Error material
func NewErrorTexture(name string) *Texture2D {
	return NewTexture(
		name,
		8,
		8,
		adapter.TextureFormatFromVtfFormat(uint32(format.RGB888)),
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

	// The first frame of every mipmap. The largest is the texture's colour.
	mipmaps := make([][]uint8, read.MipmapCount())
	for level := range mipmaps {
		if mipmaps[level], err = read.Mipmap(level, 0, 0, 0); err != nil {
			return nil, err
		}
	}

	texture := NewTexture(path,
		int(read.Header().Width),
		int(read.Header().Height),
		uint32(read.Header().HighResImageFormat),
		mipmaps[0])
	texture.mipmaps = mipmaps[1:]
	return texture, nil
}

// TextureAtlas is a simple 2d texture atlas.
// Does NOT support transparency
type TextureAtlas struct {
	rectangles []AtlasTexture
	// pages is the number of images each texture has. Each page holds one of every texture's images, at the same place
	// as in the first page, and the pages are stacked one below another.
	pages int

	populatedWidth, populatedHeight int
	width, height                   int
	colour                          []uint8
	format                          uint32
	bytesPerPixel                   int
}

func (atlas *TextureAtlas) AtlasEntry(index int) *AtlasTexture {
	if index > len(atlas.rectangles) {
		return nil
	}
	return &atlas.rectangles[index]
}

func (atlas *TextureAtlas) Format() uint32 {
	return atlas.format
}

func (atlas *TextureAtlas) Width() int {
	return atlas.width
}

func (atlas *TextureAtlas) Height() int {
	return atlas.height
}

func (atlas *TextureAtlas) PopulatedWidth() int {
	return atlas.populatedWidth
}

func (atlas *TextureAtlas) PopulatedHeight() int {
	return atlas.populatedHeight
}

// PageHeight is the height of each of the atlas's pages
func (atlas *TextureAtlas) PageHeight() int {
	return atlas.height / atlas.pages
}

func (atlas *TextureAtlas) Image() []uint8 {
	return atlas.colour
}

// Free color data from memory
func (texture *TextureAtlas) Release() {
	texture.colour = nil
	for idx := range texture.rectangles {
		texture.rectangles[idx].Release()
	}
}

// AddRaw adds a texture of RGBA colour to the atlas. colour holds the texture's image for each page, one after another.
func (atlas *TextureAtlas) AddRaw(width, height int, colour []uint8) *AtlasTexture {
	atlas.rectangles = append(atlas.rectangles, AtlasTexture{
		id:     len(atlas.rectangles),
		W:      width,
		H:      height,
		X:      0,
		Y:      0,
		colour: colour,
	})

	return &(atlas.rectangles[len(atlas.rectangles)-1])
}

func (atlas *TextureAtlas) Pack() []AtlasTexture {
	// STEP 1: GENERATE PACKED POSITIONS

	padding := 0 // total per axis (e.g 2 = 1unit each side)

	// calculate total box area and maximum box width
	area := 0
	maxWidth := 0

	maxX := 0
	maxY := 0

	for _, box := range atlas.rectangles {
		area += (box.W + padding) * (box.H + padding)
		maxWidth = int(math.Max(float64(maxWidth), float64(box.W+padding)))
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
	packed := make([]AtlasTexture, len(atlas.rectangles))

	badCounter := 0
	for idx, box := range atlas.rectangles {
		// An empty box takes no space
		if box.W == 0 || box.H == 0 {
			packed[idx] = AtlasTexture{colour: box.colour, id: box.id}
			continue
		}
		// look through spaces backwards so that we check smaller spaces first
		for i := len(spaces) - 1; i >= 0; i-- {
			space := &(spaces[i])

			// look for empty spaces that can accommodate the current box
			if box.W+padding > space.w || box.H+padding > space.h {
				continue
			}

			// found the space; add the box to its top-left corner
			// |-------|-------|
			// |  box  |       |
			// |_______|       |
			// |         space |
			// |_______________|
			if space.x == 0 && space.y == 0 {
				badCounter++
			}
			packed[idx] = AtlasTexture{W: box.W + padding, H: box.H + padding, X: float32(space.x), Y: float32(space.y), colour: box.colour, id: box.id}
			maxX = int(math.Max(float64(maxX), float64(packed[idx].X+float32(packed[idx].W))))
			maxY = int(math.Max(float64(maxY), float64(packed[idx].Y+float32(packed[idx].H))))
			// Insert colour data here to skip some duplication

			if int(box.W+padding) == space.w && int(box.H+padding) == space.h {
				// space matches the box exactly; remove it
				last := spaces[len(spaces)-1]
				spaces = spaces[:len(spaces)-1]

				if i < len(spaces) {
					spaces[i] = last
				}
			} else if box.H+padding == space.h {
				// space matches the box height; update it accordingly
				// |-------|---------------|
				// |  box  | updated space |
				// |_______|_______________|
				space.x += box.W + padding
				space.w -= box.W + padding
			} else if box.W+padding == space.w {
				// space matches the box width; update it accordingly
				// |---------------|
				// |      box      |
				// |_______________|
				// | updated space |
				// |_______________|
				space.y += box.H + padding
				space.h -= box.H + padding
			} else {
				// otherwise the box splits the space into two spaces
				// |-------|-----------|
				// |  box  | new space |
				// |_______|___________|
				// | updated space     |
				// |___________________|
				newSpace := atlasSpace{
					x: space.x + (box.W + padding),
					y: space.y,
					w: space.w - (box.W + padding),
					h: box.H + padding,
				}
				// Update the space before appending: append can move spaces, leaving space pointing at the old copy
				space.y += box.H + padding
				space.h -= box.H + padding
				spaces = append(spaces, newSpace)
			}
			break
		}
	}

	atlas.populatedWidth = maxX
	atlas.populatedHeight = maxY * atlas.pages
	atlas.width = maxX
	atlas.height = maxY * atlas.pages
	//atlas.width = int(math.Pow(2, math.Ceil(math.Log(float64(maxX))/math.Log(2))))
	//atlas.height = int(math.Pow(2, math.Ceil(math.Log(float64(maxY))/math.Log(2))))
	atlas.colour = make([]uint8, atlas.width*atlas.height*atlas.bytesPerPixel)

	// STEP 2: PACK TEXTURES
	for _, rect := range packed {
		for page := 0; page < atlas.pages; page++ {
			atlas.writeBytes(&rect, page, maxY)
		}
	}

	atlas.rectangles = nil

	// STEP 3: Restore original order so rectangles can be mapped back to faces
	sort.Slice(packed, func(i, j int) bool {
		return packed[i].id < packed[j].id
	})

	atlas.rectangles = packed

	return atlas.rectangles
}

// writeBytes writes a texture's image for a page into the atlas, whose pages are pageHeight high
func (atlas *TextureAtlas) writeBytes(rect *AtlasTexture, page int, pageHeight int) {
	// Skip rows, then indent into the baseAtlasOffset of the current row
	rowSizeInBytes := atlas.width * atlas.bytesPerPixel

	start := (rowSizeInBytes * (int(rect.Y) + page*pageHeight)) + (atlas.bytesPerPixel * int(rect.X)) // Number of rows in + number of bytes across
	image := rect.colour[page*rect.W*rect.H*4:]
	for rowY := 0; rowY < rect.H; rowY++ {
		for rowX := 0; rowX < rect.W; rowX++ {
			atlas.colour[start+(rowX*atlas.bytesPerPixel)+0] = image[(rowY*4*rect.W)+(rowX*4)+0]
			atlas.colour[start+(rowX*atlas.bytesPerPixel)+1] = image[(rowY*4*rect.W)+(rowX*4)+1]
			atlas.colour[start+(rowX*atlas.bytesPerPixel)+2] = image[(rowY*4*rect.W)+(rowX*4)+2]
			atlas.colour[start+(rowX*atlas.bytesPerPixel)+3] = 255
		}

		start += rowSizeInBytes
	}
}

func NewTextureAtlas(width, height int) *TextureAtlas {
	atlas := NewPagedTextureAtlas(1)
	atlas.width, atlas.height = width, height
	return atlas
}

// NewPagedTextureAtlas creates an atlas whose textures each have an image for every one of its pages
func NewPagedTextureAtlas(pages int) *TextureAtlas {
	return &TextureAtlas{
		pages:         pages,
		rectangles:    []AtlasTexture{},
		format:        adapter.TextureFormatFromVtfFormat(uint32(format.RGBA8888)),
		bytesPerPixel: 4,
	}
}

type AtlasTexture struct {
	id   int
	W, H int
	X, Y float32

	colour []uint8
}

func (atlasTexture *AtlasTexture) Release() {
	atlasTexture.colour = nil
}

type atlasSpace struct {
	x, y, w, h int
}
