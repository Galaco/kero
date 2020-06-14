package cache

import "github.com/galaco/kero/framework/graphics"

const (
	ErrorTexturePath = "materials/error.vtf"
)

type Texture struct {
	items map[string]*graphics.Texture2D
}

// Add
func (cache *Texture) Add(name string, item *graphics.Texture2D) {
	cache.items[name] = item
}

// Find
func (cache *Texture) Find(name string) *graphics.Texture2D {
	return cache.items[name]
}

// NewTextureCache
func NewTextureCache() *Texture {
	return &Texture{
		items: map[string]*graphics.Texture2D{},
	}
}
