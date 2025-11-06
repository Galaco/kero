package debug

import (
	"github.com/galaco/kero/framework/graphics/adapter"
	gfxdebug "github.com/galaco/kero/framework/graphics/debug"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// DebugRenderer handles rendering of debug primitives with efficient GPU resource management
type DebugRenderer struct {
	shader         *adapter.Shader
	buffer         *gfxdebug.DebugDrawBuffer
	vao            uint32
	vbo            uint32
	bufferCapacity int
	enabled        bool
}

const (
	// Initial buffer size for debug vertices (grows dynamically if needed)
	initialBufferCapacity = 10000
	// Bytes per vertex: 3 floats (position) + 3 floats (color)
	floatsPerVertex = 6
)

// NewDebugRenderer creates a new debug renderer with a dedicated shader
func NewDebugRenderer(shader *adapter.Shader, buffer *gfxdebug.DebugDrawBuffer) *DebugRenderer {
	r := &DebugRenderer{
		shader:         shader,
		buffer:         buffer,
		bufferCapacity: initialBufferCapacity,
		enabled:        true,
	}
	r.initialize()
	return r
}

// initialize sets up persistent GPU resources
func (r *DebugRenderer) initialize() {
	// Create persistent VAO and VBO
	gl.GenVertexArrays(1, &r.vao)
	gl.GenBuffers(1, &r.vbo)

	gl.BindVertexArray(r.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vbo)

	// Allocate initial buffer (will grow if needed)
	gl.BufferData(gl.ARRAY_BUFFER, r.bufferCapacity*floatsPerVertex*4, nil, gl.DYNAMIC_DRAW)

	// Setup vertex attributes
	// Location 0: Position (vec3)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, floatsPerVertex*4, nil)
	gl.EnableVertexAttribArray(0)

	// Location 1: Color (vec3)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, floatsPerVertex*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)
}

// Render draws all debug primitives in the buffer
func (r *DebugRenderer) Render(projection, view mgl32.Mat4) {
	if !r.enabled || r.buffer == nil {
		return
	}

	primitives := r.buffer.GetPrimitives()
	if len(primitives) == 0 {
		return
	}

	// Save current render state
	var depthTestEnabled bool
	var depthWriteEnabled bool
	var cullingEnabled bool
	gl.GetBooleanv(gl.DEPTH_TEST, &depthTestEnabled)
	gl.GetBooleanv(gl.DEPTH_WRITEMASK, &depthWriteEnabled)
	gl.GetBooleanv(gl.CULL_FACE, &cullingEnabled)

	// Configure debug render state
	adapter.DisableDepthTesting() // Debug lines always visible
	adapter.EnableZBufferWrite()  // Still write to depth buffer
	gl.Disable(gl.CULL_FACE)      // Show all debug geometry

	// Bind shader and set uniforms
	r.shader.Bind()
	adapter.PushMat4(r.shader.GetUniform("projection"), 1, false, projection)
	adapter.PushMat4(r.shader.GetUniform("view"), 1, false, view)

	// Batch primitives by type for efficient rendering
	r.renderPrimitivesByType(primitives, gfxdebug.Lines, gl.LINES)
	r.renderPrimitivesByType(primitives, gfxdebug.Triangles, gl.TRIANGLES)
	r.renderPrimitivesByType(primitives, gfxdebug.Points, gl.POINTS)

	// Restore previous render state
	if depthTestEnabled {
		adapter.EnableDepthTesting()
	}
	if !depthWriteEnabled {
		adapter.DisableZBufferWrite()
	}
	if cullingEnabled {
		gl.Enable(gl.CULL_FACE)
	}
}

// renderPrimitivesByType renders all primitives of a specific type in batches
func (r *DebugRenderer) renderPrimitivesByType(primitives []gfxdebug.DebugPrimitive, primType gfxdebug.PrimitiveType, glMode uint32) {
	gl.BindVertexArray(r.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vbo)

	for _, prim := range primitives {
		if prim.GetType() != primType {
			continue
		}

		vertices := prim.GetVertices()
		if len(vertices) == 0 {
			continue
		}

		// Build interleaved vertex data (position + color)
		data := r.buildVertexData(vertices, prim.GetColor())

		// Resize buffer if needed
		if len(data)/floatsPerVertex > r.bufferCapacity {
			r.bufferCapacity = len(data) / floatsPerVertex
			gl.BufferData(gl.ARRAY_BUFFER, len(data)*4, gl.Ptr(data), gl.DYNAMIC_DRAW)
		} else {
			gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(data)*4, gl.Ptr(data))
		}

		// Set model matrix for this primitive
		adapter.PushMat4(r.shader.GetUniform("model"), 1, false, prim.GetTransform())

		// Draw
		gl.DrawArrays(glMode, 0, int32(len(vertices)))
	}

	gl.BindVertexArray(0)
}

// buildVertexData converts vertices and color into interleaved format
func (r *DebugRenderer) buildVertexData(vertices []mgl32.Vec3, color mgl32.Vec3) []float32 {
	data := make([]float32, len(vertices)*floatsPerVertex)
	for i, v := range vertices {
		offset := i * floatsPerVertex
		data[offset+0] = v.X()
		data[offset+1] = v.Y()
		data[offset+2] = v.Z()
		data[offset+3] = color.X()
		data[offset+4] = color.Y()
		data[offset+5] = color.Z()
	}
	return data
}

// SetEnabled enables or disables debug rendering
func (r *DebugRenderer) SetEnabled(enabled bool) {
	r.enabled = enabled
}

// IsEnabled returns whether debug rendering is enabled
func (r *DebugRenderer) IsEnabled() bool {
	return r.enabled
}

// Cleanup releases GPU resources
func (r *DebugRenderer) Cleanup() {
	if r.vbo != 0 {
		gl.DeleteBuffers(1, &r.vbo)
		r.vbo = 0
	}
	if r.vao != 0 {
		gl.DeleteVertexArrays(1, &r.vao)
		r.vao = 0
	}
}
