package adapter

import (
	"github.com/go-gl/gl/v4.1-core/gl"
)

const (
	GBufferTextureTypePosition          = 0
	GBufferTextureTypeDiffuse           = 1
	GBufferTextureTypeNormal            = 2
	GBufferTextureTypeTextureCoordinate = 3
	gBufferTextureCount                 = 3
)

type GBuffer struct {
	fbo         uint32
	Textures    [3]uint32
	depthBuffer uint32
}

func (gbuffer *GBuffer) Initialize(width, height int) bool {
	// Create the FBO
	gl.GenFramebuffers(1, &gbuffer.fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, gbuffer.fbo)

	gl.GenTextures(1, &gbuffer.Textures[0])
	gl.BindTexture(gl.TEXTURE_2D, gbuffer.Textures[0])
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA8, int32(width), int32(height), 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, gbuffer.Textures[0], 0)
	gl.BindTexture(gl.TEXTURE_2D, 0)

	for i := uint32(1); i < 3; i++ {
		gl.GenTextures(1, &gbuffer.Textures[i])
		gl.BindTexture(gl.TEXTURE_2D, gbuffer.Textures[i])
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA16F, int32(width), int32(height), 0, gl.RGBA, gl.FLOAT, nil)
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0+i, gl.TEXTURE_2D, gbuffer.Textures[i], 0)
		gl.BindTexture(gl.TEXTURE_2D, 0)
	}

	// depth
	gl.GenRenderbuffers(1, &gbuffer.depthBuffer)
	gl.BindRenderbuffer(gl.RENDERBUFFER, gbuffer.depthBuffer)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT32, int32(width), int32(height))
	gl.BindRenderbuffer(gl.RENDERBUFFER, 0)
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, gbuffer.depthBuffer)

	drawBuffers := [3]uint32{gl.COLOR_ATTACHMENT0, gl.COLOR_ATTACHMENT1, gl.COLOR_ATTACHMENT2}
	gl.DrawBuffers(3, &drawBuffers[0])

	status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER)

	if status != gl.FRAMEBUFFER_COMPLETE {
		return false
	}

	// restore default FBO
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	return true
}

func (gbuffer *GBuffer) BindReadWrite() {
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, gbuffer.fbo)
}

func (gbuffer *GBuffer) BindReadOnly() {
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, gbuffer.fbo)
}

func (gbuffer *GBuffer) SetReadBuffer(textureType uint32) {
	gl.ReadBuffer(gl.COLOR_ATTACHMENT0 + textureType)
}
