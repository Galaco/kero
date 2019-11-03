package adapter

import (
	"github.com/go-gl/gl/v4.1-core/gl"
)

func BindFrameBuffer(id uint32) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}

func BindFrameBufferDraw(id uint32) {
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, id)
}

func BlitDepthBuffer(width, height int32) {
	gl.BlitFramebuffer(
		0,
		0,
		width,
		height,
		0,
		0,
		width,
		height,
		gl.DEPTH_BUFFER_BIT,
		gl.NEAREST)
}
