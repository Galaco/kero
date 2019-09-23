package graphics

import (
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/vtf"
	"github.com/galaco/vtf/format"
	"strings"
)

// Texture is a material defined by raw/computed colour data
type Texture struct {
	filePath string
	width    int
	height   int
	format   uint32
	colour   []uint8
}

// Format returns colour format
func (texture *Texture) Format() uint32 {
	return texture.format
}

// Width
func (texture *Texture) Width() int {
	return texture.width
}

// Height
func (texture *Texture) Height() int {
	return texture.height
}

// Image returns raw colour data
func (texture *Texture) Image() []uint8 {
	return texture.colour
}

// LoadTexture
func LoadTexture(filePath string) (*Texture, error) {
	if !strings.HasSuffix(filePath, filesystem.ExtensionVtf) {
		filePath = filePath + filesystem.ExtensionVtf
	}
	return readVtf(filesystem.BasePathMaterial + filePath)
}

func NewTexture(filePath string, width, height int, format uint32, colour []uint8) *Texture {
	return &Texture{
		filePath: filePath,
		width:    width,
		height:   height,
		format:   textureFormatFromVtfFormat(format),
		colour:   colour,
	}
}

// NewError returns new Error material
func NewErrorTexture(name string) *Texture {
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
func readVtf(path string) (*Texture, error) {
	stream, err := filesystem.Singleton().GetFile(path)
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
