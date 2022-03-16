package cache

type GpuItem struct {
	items map[string]uint32
}

// Add
func (cache *GpuItem) Add(name string, item uint32) {
	cache.items[name] = item
}

// Find
func (cache *GpuItem) Find(name string) uint32 {
	return cache.items[name]
}
func (cache *GpuItem) All() map[string]uint32 {
	return cache.items
}

// NewTextureCache
func NewGpuItemCache() GpuItem {
	return GpuItem{
		items: map[string]uint32{},
	}
}
