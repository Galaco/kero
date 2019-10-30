package adapter

import (
	"github.com/go-gl/gl/v4.1-core/gl"
)

const (
	GBufferTextureTypePosition          = 0
	GBufferTextureTypeDiffuse           = 1
	GBufferTextureTypeNormal            = 2
	GBufferTextureTypeTextureCoordinate = 3
	gBufferTextureCount                 = 4
)

type GBuffer struct {
	fbo          uint32
	textures     [gBufferTextureCount]uint32
	depthTexture uint32
}

func (gbuffer *GBuffer) Initialize(width, height int) bool {
	// Create the FBO
	gl.GenFramebuffers(1, &gbuffer.fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, gbuffer.fbo)

	// Create the gbuffer textures
	gl.GenTextures(int32(gBufferTextureCount), &gbuffer.textures[0])
	gl.GenTextures(1, &gbuffer.depthTexture)

	for i := uint32(0); i < uint32(gBufferTextureCount); i++ {
		gl.BindTexture(gl.TEXTURE_2D, gbuffer.textures[i])
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB32F, int32(width), int32(height), 0, gl.RGB, gl.FLOAT, nil)
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0+i, gl.TEXTURE_2D, gbuffer.textures[i], 0)
	}

	// depth
	gl.BindTexture(gl.TEXTURE_2D, gbuffer.depthTexture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.DEPTH_COMPONENT32F, int32(width), int32(height), 0, gl.DEPTH_COMPONENT, gl.FLOAT,
		nil)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, gbuffer.depthTexture, 0)

	drawBuffers := []uint32{gl.COLOR_ATTACHMENT0, gl.COLOR_ATTACHMENT1, gl.COLOR_ATTACHMENT2, gl.COLOR_ATTACHMENT3}
	gl.DrawBuffers(int32(len(drawBuffers)), &drawBuffers[0])

	status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER)

	if status != gl.FRAMEBUFFER_COMPLETE {
		return false
	}

	// restore default FBO
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0)

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
