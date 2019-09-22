package shaders

import (
	"github.com/galaco/gosigl"
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/systems/renderer/cache"
)

func LoadShaders() (*cache.Shader, error) {
	lightmappedGenericShader := graphics.NewShader()
	if err := lightmappedGenericShader.Add(gosigl.VertexShader, LightMappedGenericVertex); err != nil {
		return nil, err
	}
	if err := lightmappedGenericShader.Add(gosigl.FragmentShader, LightMappedGenericFragment); err != nil {
		return nil, err
	}
	lightmappedGenericShader.Finish()

	shaderCache := cache.NewShaderCache()
	shaderCache.Add("LightMappedGeneric", lightmappedGenericShader)

	return shaderCache, nil
}
