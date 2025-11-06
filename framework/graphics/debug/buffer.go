package debug

import (
	"sync"

	"github.com/go-gl/mathgl/mgl32"
)

// DebugDrawBuffer is a thread-safe collection of debug primitives
// that can be populated by various subsystems and consumed by the renderer
type DebugDrawBuffer struct {
	primitives []DebugPrimitive
	mutex      sync.RWMutex
}

// NewDebugDrawBuffer creates a new empty debug draw buffer
func NewDebugDrawBuffer() *DebugDrawBuffer {
	return &DebugDrawBuffer{
		primitives: make([]DebugPrimitive, 0, 64),
	}
}

// AddLines adds a line set to the debug buffer
func (b *DebugDrawBuffer) AddLines(vertices []mgl32.Vec3, color mgl32.Vec3, transform mgl32.Mat4) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if len(vertices) == 0 {
		return
	}

	b.primitives = append(b.primitives, NewDebugLineSet(vertices, color, transform))
}

// AddTriangles adds a triangle set to the debug buffer
func (b *DebugDrawBuffer) AddTriangles(vertices []mgl32.Vec3, color mgl32.Vec3, transform mgl32.Mat4) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if len(vertices) == 0 {
		return
	}

	b.primitives = append(b.primitives, NewDebugTriangleSet(vertices, color, transform))
}

// AddPoints adds a point set to the debug buffer
func (b *DebugDrawBuffer) AddPoints(vertices []mgl32.Vec3, color mgl32.Vec3, transform mgl32.Mat4) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if len(vertices) == 0 {
		return
	}

	b.primitives = append(b.primitives, NewDebugPointSet(vertices, color, transform))
}

// AddPrimitive adds a custom debug primitive to the buffer
func (b *DebugDrawBuffer) AddPrimitive(primitive DebugPrimitive) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if primitive == nil {
		return
	}

	b.primitives = append(b.primitives, primitive)
}

// GetPrimitives returns a copy of all primitives in the buffer
// This is safe to call from the render thread while other threads add primitives
func (b *DebugDrawBuffer) GetPrimitives() []DebugPrimitive {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	// Return a copy to avoid concurrent modification issues
	result := make([]DebugPrimitive, len(b.primitives))
	copy(result, b.primitives)
	return result
}

// Clear removes all primitives from the buffer
// This should be called after rendering is complete
func (b *DebugDrawBuffer) Clear() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// Reuse the slice to avoid allocations
	b.primitives = b.primitives[:0]
}

// Count returns the number of primitives in the buffer
func (b *DebugDrawBuffer) Count() int {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	return len(b.primitives)
}
