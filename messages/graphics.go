package messages

import (
	"github.com/galaco/kero/event/message"
	"github.com/galaco/kero/framework/graphics"
)

const (
	TypeTextureLoaded = message.Type("TextureLoaded")
	TypeMaterialLoaded = message.Type("MaterialLoaded")
)

type TextureLoaded struct {
	texture *graphics.Texture
}

func (msg *TextureLoaded) Type() message.Type {
	return TypeTextureLoaded
}

func (msg *TextureLoaded) Texture() *graphics.Texture {
	return msg.texture
}

func NewTextureLoaded(texture *graphics.Texture) *TextureLoaded {
	return &TextureLoaded{
		texture: texture,
	}
}


type MaterialLoaded struct {
	material *graphics.Material
}

func (msg *MaterialLoaded) Type() message.Type {
	return TypeMaterialLoaded
}

func (msg *MaterialLoaded) Material() *graphics.Material {
	return msg.material
}

func NewMaterialLoaded(material *graphics.Material) *MaterialLoaded {
	return &MaterialLoaded{
		material: material,
	}
}

