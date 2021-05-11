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

func GpuError() error {
	if glError := gl.GetError(); glError != gl.NO_ERROR {
		return fmt.Errorf("gl error. Code: %d", glError)
	}
	return nil
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

	drawLinesInternal(points)
}

func DrawDebugLines(points []float32, color mgl32.Vec3) {
	if len(points) == 0 {
		return
	}
	// Vertex data
	combinedPoints := make([]float32, 0, len(points) * 2)

	// Unpleasant but masks the data format from adapter users
	for i := 0; i < len(points); i += 3 {
		combinedPoints = append(combinedPoints, points[i], points[i+1], points[i+2], color.X(), color.Y(), color.Z())
	}

	drawLinesInternal(combinedPoints)
}

func drawLinesInternal(points []float32) {
	gl.DeleteBuffers(1, &drawLineVBO)
	gl.DeleteVertexArrays(1, &drawLineVAO)
	gl.GenBuffers(1, &drawLineVBO)
	gl.GenVertexArrays(1, &drawLineVAO)
	gl.BindVertexArray(drawLineVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, drawLineVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(points) * 3, gl.Ptr(points), gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 6 * 4, nil)
	gl.EnableVertexAttribArray(1)
	// gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6 * 4, (GLvoid*)(3 * sizeof(GLfloat)))
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6 * 4, nil)
	gl.BindVertexArray(0)

	gl.BindVertexArray(drawLineVAO)
	gl.DrawArrays(gl.LINES, 0, int32(len(points) / 6))
	gl.BindVertexArray(0)
}