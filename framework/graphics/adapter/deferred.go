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
	PositionBuffer, NormalBuffer, AlbedoSpecularBuffer    uint32
	depthBuffer uint32
}

func (gbuffer *GBuffer) Initialize(width, height int) bool {
	// Create the FBO
	gl.GenFramebuffers(1, &gbuffer.fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, gbuffer.fbo)

	// Position
	gl.GenTextures(1, &gbuffer.PositionBuffer)
	gl.BindTexture(gl.TEXTURE_2D, gbuffer.PositionBuffer)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB16F, int32(width), int32(height), 0, gl.RGB, gl.FLOAT, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, gbuffer.PositionBuffer, 0)

	// Normal
	gl.GenTextures(1, &gbuffer.NormalBuffer)
	gl.BindTexture(gl.TEXTURE_2D, gbuffer.NormalBuffer)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB16F, int32(width), int32(height), 0, gl.RGB, gl.FLOAT, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT1, gl.TEXTURE_2D, gbuffer.NormalBuffer, 0)

	// Albedo
	gl.GenTextures(1, &gbuffer.AlbedoSpecularBuffer)
	gl.BindTexture(gl.TEXTURE_2D, gbuffer.AlbedoSpecularBuffer)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(width), int32(height), 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT2, gl.TEXTURE_2D, gbuffer.AlbedoSpecularBuffer, 0)

	drawBuffers := [3]uint32{gl.COLOR_ATTACHMENT0, gl.COLOR_ATTACHMENT1, gl.COLOR_ATTACHMENT2}
	gl.DrawBuffers(3, &drawBuffers[0])

	// depth
	gl.GenRenderbuffers(1, &gbuffer.depthBuffer)
	gl.BindRenderbuffer(gl.RENDERBUFFER, gbuffer.depthBuffer)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT, int32(width), int32(height))
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, gbuffer.depthBuffer)

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
