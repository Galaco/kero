package utils

import (
	"github.com/galaco/kero/framework/graphics"
	"image"
	"image/png"
	"os"
)

func DumpLightmap(name string, im graphics.Texture) {
	remapped := make([]uint8, (len(im.Image())/3)*4)

	for i := 0; i < len(im.Image())/3; i++ {
		remapped[(i*4)+0] = im.Image()[(i*3)+0]
		remapped[(i*4)+1] = im.Image()[(i*3)+1]
		remapped[(i*4)+2] = im.Image()[(i*3)+2]
		remapped[(i*4)+3] = 255
	}

	img := image.NewRGBA(image.Rect(0, 0, im.Width(), im.Height()))
	copy(img.Pix, remapped)

	outfile, _ := os.Create("./" + name + ".jpg")
	defer outfile.Close()
	png.Encode(outfile, img)
}
