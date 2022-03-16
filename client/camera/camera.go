package camera

import (
	"github.com/galaco/kero/internal/framework/graphics"
	"github.com/galaco/kero/internal/framework/input"
)

type Camera struct {
	boundCamera *graphics.Camera
}

func (cam *Camera) Rotate(x, y, z float32) {
	cam.boundCamera.Rotate(x, y, z)
}

func (cam *Camera) Update(dt float64) {
	if input.Keyboard().IsKeyPressed(input.KeyW) {
		cam.boundCamera.Forwards(dt)
	}
	if input.Keyboard().IsKeyPressed(input.KeyA) {
		cam.boundCamera.Left(dt)
	}
	if input.Keyboard().IsKeyPressed(input.KeyS) {
		cam.boundCamera.Backwards(dt)
	}
	if input.Keyboard().IsKeyPressed(input.KeyD) {
		cam.boundCamera.Right(dt)
	}
}

func NewCamera(boundCamera *graphics.Camera) *Camera {
	return &Camera{boundCamera: boundCamera}
}
