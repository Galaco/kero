package utils

import (
	"image"
	"image/png"
	"os"

	"github.com/galaco/kero/internal/framework/graphics"
)

// DumpLightmap exports the loaded lightmap texture atlas as a JPG
func DumpLightmap(name string, im graphics.Texture) {
	img := image.NewRGBA(image.Rect(0, 0, im.Width(), im.Height()))
	copy(img.Pix, im.Image())

	outfile, _ := os.Create("./" + name + ".jpg")
	defer outfile.Close()
	png.Encode(outfile, img)
}
