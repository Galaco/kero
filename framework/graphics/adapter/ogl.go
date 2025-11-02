package adapter

import (
	"fmt"
	"github.com/galaco/gosigl"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	ShaderTypeVertex   = gosigl.VertexShader
	ShaderTypeFragment = gosigl.FragmentShader
)

type Texture interface {
	Format() uint32
	Width() int
	Height() int
	Image() []uint8
	Release()
}

type Mesh interface {
	Vertices() []float32
	Normals() []float32
	UVs() []float32
	Tangents() []float32
	LightmapUVs() []float32
	Indices() []uint32
}

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

func ClearDepthBuffer() {
	gl.Clear(gl.DEPTH_BUFFER_BIT)
}

func UploadTexture(texture Texture) uint32 {
	return uint32(gosigl.CreateTexture2D(
		gosigl.TextureSlot(0),
		texture.Width(),
		texture.Height(),
		texture.Image(),
		gosigl.PixelFormat(texture.Format()),
		false))
}

func ReleaseTextureResource(texture Texture) {
	texture.Release()
}

func DeleteTextureResource(textureId uint32) {
	gosigl.DeleteTextures(gosigl.TextureBindingId(textureId))
}

func DeleteMeshResource(mesh GpuMesh) {
	gosigl.DeleteMesh(mesh)
}

func UploadLightmap(texture Texture) uint32 {
	textureBuffer := uint32(0)
	gl.GenTextures(1, &textureBuffer)
	gl.ActiveTexture(gl.TEXTURE4)
	gl.BindTexture(gl.TEXTURE_2D, textureBuffer)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(texture.Width()),
		int32(texture.Height()),
		0,
		texture.Format(),
		gl.UNSIGNED_BYTE,
		gl.Ptr(texture.Image()))

	return textureBuffer
}

func UploadCubemap(textures []Texture) uint32 {
	colour := [6][]byte{
		textures[0].Image(),
		textures[1].Image(),
		textures[2].Image(),
		textures[3].Image(),
		textures[4].Image(),
		textures[5].Image(),
	}

	return uint32(gosigl.CreateTextureCubemap(
		gosigl.TextureSlot(0),
		textures[0].Width(),
		textures[0].Height(),
		colour,
		gosigl.PixelFormat(textures[0].Format()),
		true))
}

func BindTexture(id uint32) {
	gosigl.BindTexture2D(gosigl.TextureSlot(0), gosigl.TextureBindingId(id))
}

func BindLightmap(id uint32) {
	gosigl.BindTexture2D(gosigl.TextureSlot(4), gosigl.TextureBindingId(id))
}

func BindCubemap(id uint32) {
	gosigl.BindTextureCubemap(gosigl.TextureSlot(0), gosigl.TextureBindingId(id))
}

// textureFormatFromVtfFormat swap vtf format to openGL format
func TextureFormatFromVtfFormat(vtfFormat uint32) uint32 {
	switch vtfFormat {
	case 0:
		return gl.RGBA
	case 2:
		return gl.RGB
	case 3:
		return gl.BGR
	case 12:
		return gl.BGRA
	case 13:
		return gl.COMPRESSED_RGB_S3TC_DXT1_EXT
	case 14:
		return gl.COMPRESSED_RGBA_S3TC_DXT3_EXT
	case 15:
		return gl.COMPRESSED_RGBA_S3TC_DXT5_EXT
	default:
		return gl.RGB
	}
}

type GpuMesh *gosigl.VertexObject

func UploadMesh(mesh Mesh) GpuMesh {
	gpuResource := gosigl.NewMesh(mesh.Vertices())
	gosigl.CreateVertexAttribute(gpuResource, mesh.UVs(), 2)
	gosigl.CreateVertexAttribute(gpuResource, mesh.Normals(), 3)
	gosigl.CreateVertexAttribute(gpuResource, mesh.Tangents(), 4)
	if mesh.LightmapUVs() == nil {
		defaultUVs := make([]float32, len(mesh.UVs()))
		for i := range defaultUVs {
			defaultUVs[i] = -1
		}
		gosigl.CreateVertexAttribute(gpuResource, defaultUVs, 2)
	} else {
		gosigl.CreateVertexAttribute(gpuResource, mesh.LightmapUVs(), 2)
	}

	if len(mesh.Indices()) > 0 {
		gosigl.SetElementArrayAttribute(gpuResource, mesh.Indices())
	}

	gosigl.FinishMesh()

	return gpuResource
}

func DrawArray(offset int, num int) {
	gosigl.DrawArray(offset, num)
}

func DrawIndexedArray(num int, offset int, indices []uint32) {
	gosigl.DrawElements(num, offset, indices)
}

// DrawMultiIndexedArray draws multiple ranges from the same index buffer
// offsets and counts must be the same length
// This avoids index buffer uploads by using offsets into the existing buffer
func DrawMultiIndexedArray(counts []int32, offsets []int) {
	if len(counts) == 0 || len(counts) != len(offsets) {
		return
	}

	// Draw each range using glDrawElements with byte offset into the index buffer
	for i := range counts {
		// offset in bytes = offset in indices * 4 (sizeof uint32)
		gl.DrawElements(gl.TRIANGLES, counts[i], gl.UNSIGNED_INT, gl.PtrOffset(offsets[i]*4))
	}
}

func UpdateIndexArrayBuffer(indices []uint32) {
	gl.BufferSubData(gl.ELEMENT_ARRAY_BUFFER, 0, len(indices)*4, gl.Ptr(indices))
}

func DrawFace(offset int, num int, textureId uint32) {
	BindTexture(textureId)
	DrawArray(offset, num)
}

func BindMesh(mesh *GpuMesh) {
	gosigl.BindMesh(*mesh)
}

func PushMat4(uniform int32, count int, transpose bool, mat mgl32.Mat4) {
	gl.UniformMatrix4fv(uniform, int32(count), transpose, &mat[0])
}

func PushInt32(uniform int32, value int32) {
	gl.Uniform1i(uniform, value)
}

func PushFloat32(uniform int32, value float32) {
	gl.Uniform1f(uniform, value)
}

func PushVec3(uniform int32, vec mgl32.Vec3) {
	gl.Uniform3f(uniform, vec.X(), vec.Y(), vec.Z())
}

func GpuError() error {
	if glError := gl.GetError(); glError != gl.NO_ERROR {
		return fmt.Errorf("gl error. Code: %d", glError)
	}
	return nil
}

// CreateEmptyInstanceBuffer creates a persistent GPU buffer for instance data
// Buffer is allocated but not filled (use UpdateInstanceBuffer to fill)
// floatsPerInstance should be 18 (16 for mat4 + 2 for fade min/max)
func CreateEmptyInstanceBuffer(maxInstances int, floatsPerInstance int) uint32 {
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)

	// Allocate buffer with max capacity (no data uploaded yet)
	sizeInBytes := maxInstances * floatsPerInstance * 4 // 4 bytes per float
	gl.BufferData(gl.ARRAY_BUFFER, sizeInBytes, nil, gl.DYNAMIC_DRAW) // nil = allocate only

	return vbo
}

// UpdateInstanceBuffer updates an existing instance buffer with new instance data
// More efficient than recreating the buffer each frame
// data is a flat array: [mat4_1(16 floats), fade_1(2 floats), mat4_2(16 floats), fade_2(2 floats), ...]
func UpdateInstanceBuffer(vbo uint32, data []float32) {
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)

	// Upload only the visible instances (partial buffer update)
	sizeInBytes := len(data) * 4
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, sizeInBytes, gl.Ptr(data))
}

// SetupInstanceAttributes configures vertex attributes for instanced rendering with fade
// Must be called after BindMesh, before drawing
func SetupInstanceAttributes(instanceVBO uint32) {
	gl.BindBuffer(gl.ARRAY_BUFFER, instanceVBO)

	stride := 18 * 4 // 72 bytes (16 floats for mat4 + 2 floats for fade min/max)

	// mat4 model matrix (locations 5-8, one vec4 per column)
	for i := uint32(0); i < 4; i++ {
		loc := 5 + i
		gl.EnableVertexAttribArray(loc)
		gl.VertexAttribPointer(loc, 4, gl.FLOAT, false, int32(stride), gl.PtrOffset(int(i)*4*4))
		gl.VertexAttribDivisor(loc, 1) // Advance per instance, not per vertex
	}

	// vec2 fade min/max (location 9)
	gl.EnableVertexAttribArray(9)
	gl.VertexAttribPointer(9, 2, gl.FLOAT, false, int32(stride), gl.PtrOffset(16*4))
	gl.VertexAttribDivisor(9, 1) // Advance per instance
}

// DisableInstanceAttributes disables instance attributes to allow non-instanced rendering
// Call this before switching back to non-instanced rendering with the same mesh
func DisableInstanceAttributes() {
	// Disable instance attributes (locations 5-9)
	for i := uint32(5); i <= 9; i++ {
		gl.DisableVertexAttribArray(i)
		gl.VertexAttribDivisor(i, 0) // Reset divisor to 0 (per-vertex, not per-instance)
	}
}

// DrawIndexedArrayInstanced draws mesh multiple times with different matrices
func DrawIndexedArrayInstanced(indexCount int, instanceCount int) {
	gl.DrawElementsInstanced(gl.TRIANGLES, int32(indexCount), gl.UNSIGNED_INT, nil, int32(instanceCount))
}

// DeleteInstanceBuffer cleans up instance buffer
func DeleteInstanceBuffer(vbo uint32) {
	gl.DeleteBuffers(1, &vbo)
}

func EnableBlending() {
	gosigl.EnableBlend()
}

func DisableBlending() {
	gosigl.DisableBlend()
}

func EnableDepthTesting() {
	gosigl.EnableDepthTest()
}

func DisableDepthTesting() {
	gosigl.DisableDepthTest()
}

func EnableZBufferWrite() {
	gl.DepthMask(true)
}

func DisableZBufferWrite() {
	gl.DepthMask(false)
}

func EnableBackFaceCulling() {
	gosigl.EnableCullFace(gosigl.Back, gosigl.WindingClockwise)
}

func EnableFrontFaceCulling() {
	gosigl.EnableCullFace(gosigl.Front, gosigl.WindingClockwise)
}

// not a great implementation, but isolates gl specifics from outside of the adapter
var drawLineVBO, drawLineVAO uint32

func DrawLine(start, end, color mgl32.Vec3) {
	// Vertex data
	points := []float32{
		start.X(),
		start.Y(),
		start.Z(),
		color.X(),
		color.Y(),
		color.Z(),
		end.X(),
		end.Y(),
		end.Z(),
		color.X(),
		color.Y(),
		color.Z(),
	}

	drawCommonInternal(points, gl.LINES)
}

func DrawDebugLines(points []float32, color mgl32.Vec3) {
	if len(points) == 0 {
		return
	}
	// Vertex data
	combinedPoints := make([]float32, 0, len(points)*2)

	// Unpleasant but masks the data format from adapter users
	for i := 0; i < len(points); i += 3 {
		combinedPoints = append(combinedPoints, points[i], points[i+1], points[i+2], color.X(), color.Y(), color.Z())
	}

	drawCommonInternal(combinedPoints, gl.LINES)
}

func DrawDebugTriangles(points []float32, color mgl32.Vec3) {
	if len(points) == 0 {
		return
	}
	// Vertex data
	combinedPoints := make([]float32, 0, len(points)*2)

	// Unpleasant but masks the data format from adapter users
	for i := 0; i < len(points); i += 3 {
		combinedPoints = append(combinedPoints, points[i], points[i+1], points[i+2], color.X(), color.Y(), color.Z())
	}

	drawCommonInternal(combinedPoints, gl.TRIANGLES)
}

func drawCommonInternal(points []float32, drawType uint32) {
	gl.DeleteBuffers(1, &drawLineVBO)
	gl.DeleteVertexArrays(1, &drawLineVAO)
	gl.GenBuffers(1, &drawLineVBO)
	gl.GenVertexArrays(1, &drawLineVAO)
	gl.BindVertexArray(drawLineVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, drawLineVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(points)*3, gl.Ptr(points), gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 6*4, nil)
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6*4, nil)
	gl.BindVertexArray(0)

	gl.BindVertexArray(drawLineVAO)
	gl.DrawArrays(drawType, 0, int32(len(points)/6))
	gl.BindVertexArray(0)
}
