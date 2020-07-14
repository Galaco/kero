package cache

import (
	"github.com/galaco/kero/framework/graphics/adapter"
)

type Shader struct {
	items map[string]*adapter.Shader
}

// Add
func (cache *Shader) Add(name string, item *adapter.Shader) {
	cache.items[name] = item
}

// Find
func (cache *Shader) Find(name string) *adapter.Shader {
	return cache.items[name]
}

// NewTextureCache
func NewShaderCache() *Shader {
	return &Shader{
		items: map[string]*adapter.Shader{},
	}
}
