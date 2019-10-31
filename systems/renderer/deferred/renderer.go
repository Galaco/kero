package deferred

import (
	"errors"
	"github.com/galaco/gosigl"
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/framework/graphics/adapter"
)

type Renderer struct {
	gbuffer       adapter.GBuffer
	width, height int32

	geometryShader *graphics.Shader
	directionalLightShader *graphics.Shader
}

func (renderer *Renderer) Init(width, height int) error {
	renderer.width = int32(width)
	renderer.height = int32(height)
	success := renderer.gbuffer.Initialize(width, height)
	if !success {
		return errors.New("failed to initialize gbuffer")
	}

	renderer.geometryShader = graphics.NewShader()
	if err := renderer.geometryShader.Add(gosigl.VertexShader, GeometryPassVertex); err != nil {
		return err
	}
	if err := renderer.geometryShader.Add(gosigl.FragmentShader, GeometryPassFragment); err != nil {
		return err
	}
	renderer.geometryShader.Finish()
	renderer.geometryShader.Bind()
	adapter.PushInt32(renderer.geometryShader.GetUniform("albedoSampler"), 0)

	renderer.directionalLightShader = graphics.NewShader()
	if err := renderer.directionalLightShader.Add(gosigl.VertexShader, DirectionalLightPassVertex); err != nil {
		return err
	}
	if err := renderer.directionalLightShader.Add(gosigl.FragmentShader, DirectionalLightPassFragment); err != nil {
		return err
	}
	renderer.directionalLightShader.Finish()
	return nil
}