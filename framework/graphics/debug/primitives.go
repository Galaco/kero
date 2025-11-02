package debug

import (
	"github.com/go-gl/mathgl/mgl32"
)

// PrimitiveType defines the type of debug primitive to render
type PrimitiveType int

const (
	// Lines renders vertices as line segments (pairs of vertices)
	Lines PrimitiveType = iota
	// Triangles renders vertices as filled triangles
	Triangles
	// Points renders vertices as points
	Points
)

// DebugPrimitive represents a renderable debug geometry
type DebugPrimitive interface {
	// GetVertices returns the vertex positions as a flat array
	GetVertices() []mgl32.Vec3
	// GetColor returns the color for all vertices in this primitive
	GetColor() mgl32.Vec3
	// GetTransform returns the model transformation matrix
	GetTransform() mgl32.Mat4
	// GetType returns the primitive type for rendering
	GetType() PrimitiveType
}

// DebugLineSet represents a collection of lines with a single color and transform
type DebugLineSet struct {
	vertices  []mgl32.Vec3
	color     mgl32.Vec3
	transform mgl32.Mat4
}

// NewDebugLineSet creates a new debug line set
func NewDebugLineSet(vertices []mgl32.Vec3, color mgl32.Vec3, transform mgl32.Mat4) *DebugLineSet {
	return &DebugLineSet{
		vertices:  vertices,
		color:     color,
		transform: transform,
	}
}

func (d *DebugLineSet) GetVertices() []mgl32.Vec3 {
	return d.vertices
}

func (d *DebugLineSet) GetColor() mgl32.Vec3 {
	return d.color
}

func (d *DebugLineSet) GetTransform() mgl32.Mat4 {
	return d.transform
}

func (d *DebugLineSet) GetType() PrimitiveType {
	return Lines
}

// DebugTriangleSet represents a collection of triangles with a single color and transform
type DebugTriangleSet struct {
	vertices  []mgl32.Vec3
	color     mgl32.Vec3
	transform mgl32.Mat4
}

// NewDebugTriangleSet creates a new debug triangle set
func NewDebugTriangleSet(vertices []mgl32.Vec3, color mgl32.Vec3, transform mgl32.Mat4) *DebugTriangleSet {
	return &DebugTriangleSet{
		vertices:  vertices,
		color:     color,
		transform: transform,
	}
}

func (d *DebugTriangleSet) GetVertices() []mgl32.Vec3 {
	return d.vertices
}

func (d *DebugTriangleSet) GetColor() mgl32.Vec3 {
	return d.color
}

func (d *DebugTriangleSet) GetTransform() mgl32.Mat4 {
	return d.transform
}

func (d *DebugTriangleSet) GetType() PrimitiveType {
	return Triangles
}

// DebugPointSet represents a collection of points with a single color and transform
type DebugPointSet struct {
	vertices  []mgl32.Vec3
	color     mgl32.Vec3
	transform mgl32.Mat4
}

// NewDebugPointSet creates a new debug point set
func NewDebugPointSet(vertices []mgl32.Vec3, color mgl32.Vec3, transform mgl32.Mat4) *DebugPointSet {
	return &DebugPointSet{
		vertices:  vertices,
		color:     color,
		transform: transform,
	}
}

func (d *DebugPointSet) GetVertices() []mgl32.Vec3 {
	return d.vertices
}

func (d *DebugPointSet) GetColor() mgl32.Vec3 {
	return d.color
}

func (d *DebugPointSet) GetTransform() mgl32.Mat4 {
	return d.transform
}

func (d *DebugPointSet) GetType() PrimitiveType {
	return Points
}