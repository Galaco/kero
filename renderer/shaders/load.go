package shaders

import (
	"github.com/galaco/gosigl"
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/renderer/cache"
)

func LoadShaders() (*cache.Shader, error) {
	shaderCache := cache.NewShaderCache()

	lightmappedGenericShader := graphics.NewShader()
	if err := lightmappedGenericShader.Add(gosigl.VertexShader, LightMappedGenericVertex); err != nil {
		return nil, err
	}
	if err := lightmappedGenericShader.Add(gosigl.FragmentShader, LightMappedGenericFragment); err != nil {
		return nil, err
	}
	lightmappedGenericShader.Finish()
	shaderCache.Add("LightMappedGeneric", lightmappedGenericShader)

	skyboxShader := graphics.NewShader()
	if err := skyboxShader.Add(gosigl.VertexShader, SkyboxVertex); err != nil {
		return nil, err
	}
	if err := skyboxShader.Add(gosigl.FragmentShader, SkyboxFragment); err != nil {
		return nil, err
	}
	skyboxShader.Finish()
	shaderCache.Add("Skybox", skyboxShader)

	return shaderCache, nil
}
