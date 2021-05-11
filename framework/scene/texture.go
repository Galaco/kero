package scene

import "github.com/galaco/kero/framework/graphics"

const (
	ErrorTexturePath    = "materials/error.vtf"
	LightmapTexturePath = "__lightmap__"
)

type TextureCache struct {
	items map[string]graphics.Texture
}

// Add
func (cache *TextureCache) Add(name string, item graphics.Texture) {
	cache.items[name] = item
}

// Find
func (cache *TextureCache) Find(name string) graphics.Texture {
	return cache.items[name]
}

func (cache *TextureCache) All() map[string]graphics.Texture {
	return cache.items
}

// NewTextureCache
func NewTextureCache() TextureCache {
	return TextureCache{
		items: map[string]graphics.Texture{},
	}
}
