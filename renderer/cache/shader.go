package cache

import (
	"github.com/galaco/kero/framework/graphics"
)

type Shader struct {
	items map[string]*graphics.Shader
}

// Add
func (cache *Shader) Add(name string, item *graphics.Shader) {
	cache.items[name] = item
}

// Find
func (cache *Shader) Find(name string) *graphics.Shader {
	return cache.items[name]
}

// NewTextureCache
func NewShaderCache() *Shader {
	return &Shader{
		items: map[string]*graphics.Shader{},
	}
}
