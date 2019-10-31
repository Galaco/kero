package deferred

import (
	graphics3d "github.com/galaco/kero/framework/graphics/3d"
	"github.com/galaco/kero/framework/graphics/adapter"
)

func (renderer *Renderer) GeometryPass(camera *graphics3d.Camera) {
	renderer.gbuffer.BindReadWrite()
	adapter.ClearColor(0,1,0,1)
	adapter.ClearAll()

	renderer.geometryShader.Bind()

	adapter.PushMat4(renderer.geometryShader.GetUniform("projection"), 1, false, camera.ProjectionMatrix())
	adapter.PushMat4(renderer.geometryShader.GetUniform("view"), 1, false, camera.ViewMatrix())
	adapter.PushMat4(renderer.geometryShader.GetUniform("model"), 1, false, camera.ModelMatrix())
	adapter.PushInt32(renderer.geometryShader.GetUniform("albedoSampler"), 0)
}