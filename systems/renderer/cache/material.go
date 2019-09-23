package cache

const (
	ErrorMaterialPath = "materials/error.vmt"
)

type GpuMaterial struct {
	Diffuse uint32
}

func NewGpuMaterial(diffuse uint32) *GpuMaterial {
	return &GpuMaterial{
		Diffuse: diffuse,
	}
}

type Material struct {
	items map[string]*GpuMaterial
}

// Add
func (cache *Material) Add(name string, item *GpuMaterial) {
	cache.items[name] = item
}

// Find
func (cache *Material) Find(name string) *GpuMaterial {
	return cache.items[name]
}

// NewTextureCache
func NewMaterialCache() *Material {
	return &Material{
		items: map[string]*GpuMaterial{},
	}
}
