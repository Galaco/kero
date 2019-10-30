package adapter

import (
	"fmt"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

func Init() error {
	return gl.Init()
}

func Viewport(x, y, width, height int32) {
	gl.Viewport(x, y, width, height)
}

func ClearColor(r, g, b, a float32) {
	gl.ClearColor(r, g, b, a)
}

func ClearAll() {
	Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func Clear(mask uint32) {
	gl.Clear(mask)
}
func PushMat4(uniform int32, count int, transpose bool, mat mgl32.Mat4) {
	gl.UniformMatrix4fv(uniform, int32(count), transpose, &mat[0])
}

func PushInt32(uniform int32, value int32) {
	gl.Uniform1i(uniform, value)
}

func GpuError() error {
	if glError := gl.GetError(); glError != gl.NO_ERROR {
		return fmt.Errorf("gl error. Code: %d", glError)
	}
	return nil
}
