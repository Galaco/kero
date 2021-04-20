package shaders

import (
	"github.com/galaco/kero/framework/graphics/adapter"
	"github.com/galaco/kero/renderer/cache"
)

func LoadShaders() (*cache.Shader, error) {
	shaderCache := cache.NewShaderCache()

	lightmappedGenericShader := adapter.NewShader()
	if err := lightmappedGenericShader.Add(adapter.ShaderTypeVertex, LightMappedGenericVertex); err != nil {
		return nil, err
	}
	if err := lightmappedGenericShader.Add(adapter.ShaderTypeFragment, LightMappedGenericFragment); err != nil {
		return nil, err
	}
	lightmappedGenericShader.Finish()
	shaderCache.Add("LightMappedGeneric", lightmappedGenericShader)

	skyboxShader := adapter.NewShader()
	if err := skyboxShader.Add(adapter.ShaderTypeVertex, SkyboxVertex); err != nil {
		return nil, err
	}
	if err := skyboxShader.Add(adapter.ShaderTypeFragment, SkyboxFragment); err != nil {
		return nil, err
	}
	skyboxShader.Finish()
	shaderCache.Add("Skybox", skyboxShader)

	return shaderCache, nil
}
