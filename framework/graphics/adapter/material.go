package adapter

import (
	"github.com/galaco/gosigl"
	"github.com/go-gl/gl/v4.1-core/gl"
)

type Texture interface {
	Width() int
	Height() int
	Format() uint32
	Image() []uint8
}

// TextureFormatFromVtfFormat swap vtf format to openGL format
func TextureFormatFromVtfFormat(vtfFormat uint32) uint32 {
	switch vtfFormat {
	case 0:
		return gl.RGBA
	case 2:
		return gl.RGB
	case 3:
		return gl.BGR
	case 12:
		return gl.BGRA
	case 13:
		return gl.COMPRESSED_RGB_S3TC_DXT1_EXT
	case 14:
		return gl.COMPRESSED_RGBA_S3TC_DXT3_EXT
	case 15:
		return gl.COMPRESSED_RGBA_S3TC_DXT5_EXT
	default:
		return gl.RGB
	}
}

func UploadTexture(texture Texture) uint32 {
	return uint32(gosigl.CreateTexture2D(
		gosigl.TextureSlot(0),
		texture.Width(),
		texture.Height(),
		texture.Image(),
		gosigl.PixelFormat(texture.Format()),
		false))
}

func UploadCubemap(textures []Texture) uint32 {
	colour := [6][]byte{
		textures[0].Image(),
		textures[1].Image(),
		textures[2].Image(),
		textures[3].Image(),
		textures[4].Image(),
		textures[5].Image(),
	}

	return uint32(gosigl.CreateTextureCubemap(
		gosigl.TextureSlot(0),
		textures[0].Width(),
		textures[0].Height(),
		colour,
		gosigl.PixelFormat(textures[0].Format()),
		true))
}

func BindTexture(id uint32) {
	gosigl.BindTexture2D(gosigl.TextureSlot(0), gosigl.TextureBindingId(id))
}

func BindCubemap(id uint32) {
	gosigl.BindTextureCubemap(gosigl.TextureSlot(0), gosigl.TextureBindingId(id))
}
