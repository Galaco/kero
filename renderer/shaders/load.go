package shaders

import (
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/renderer/cache"
)

func LoadShaders() (*cache.Shader, error) {
	shaderCache := cache.NewShaderCache()

	lightmappedGenericShader := graphics.NewShader()
	if err := lightmappedGenericShader.Add(graphics.ShaderTypeVertex, LightMappedGenericVertex); err != nil {
		return nil, err
	}
	if err := lightmappedGenericShader.Add(graphics.ShaderTypeFragment, LightMappedGenericFragment); err != nil {
		return nil, err
	}
	lightmappedGenericShader.Finish()
	shaderCache.Add("LightMappedGeneric", lightmappedGenericShader)

	skyboxShader := graphics.NewShader()
	if err := skyboxShader.Add(graphics.ShaderTypeVertex, SkyboxVertex); err != nil {
		return nil, err
	}
	if err := skyboxShader.Add(graphics.ShaderTypeFragment, SkyboxFragment); err != nil {
		return nil, err
	}
	skyboxShader.Finish()
	shaderCache.Add("Skybox", skyboxShader)

	return shaderCache, nil
}
