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

	// Instanced version of LightMappedGeneric for static props
	lightmappedGenericInstancedShader := adapter.NewShader()
	if err := lightmappedGenericInstancedShader.Add(adapter.ShaderTypeVertex, LightMappedGenericInstancedVertex); err != nil {
		return nil, err
	}
	if err := lightmappedGenericInstancedShader.Add(adapter.ShaderTypeFragment, LightMappedGenericInstancedFragment); err != nil {
		return nil, err
	}
	lightmappedGenericInstancedShader.Finish()
	shaderCache.Add("LightMappedGenericInstanced", lightmappedGenericInstancedShader)

	// WorldVertexTransition shader for 2-texture displacement blending
	worldVertexTransitionShader := adapter.NewShader()
	if err := worldVertexTransitionShader.Add(adapter.ShaderTypeVertex, WorldVertexTransitionVertex); err != nil {
		return nil, err
	}
	if err := worldVertexTransitionShader.Add(adapter.ShaderTypeFragment, WorldVertexTransitionFragment); err != nil {
		return nil, err
	}
	worldVertexTransitionShader.Finish()
	shaderCache.Add("WorldVertexTransition", worldVertexTransitionShader)

	// Debug shader for debug primitives (physics, vis, etc.)
	debugShader := adapter.NewShader()
	if err := debugShader.Add(adapter.ShaderTypeVertex, DebugVertex); err != nil {
		return nil, err
	}
	if err := debugShader.Add(adapter.ShaderTypeFragment, DebugFragment); err != nil {
		return nil, err
	}
	debugShader.Finish()
	shaderCache.Add("Debug", debugShader)

	return shaderCache, nil
}
