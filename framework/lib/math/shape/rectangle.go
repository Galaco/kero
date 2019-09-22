package shape

import "github.com/go-gl/mathgl/mgl32"

// Rect
type Rect struct {
	// Mins
	Mins mgl32.Vec2
	// Maxs
	Maxs mgl32.Vec2
}

// X
func (rect *Rect) X() float32 {
	return rect.Mins.X()
}

// Y
func (rect *Rect) Y() float32 {
	return rect.Mins.Y()
}

// Width
func (rect *Rect) Width() float32 {
	return rect.Maxs.X()
}

// Height
func (rect *Rect) Height() float32 {
	return rect.Maxs.Y()
}

// NewRect
func NewRect(mins mgl32.Vec2, maxs mgl32.Vec2) *Rect {
	return &Rect{
		Mins: mins,
		Maxs: maxs,
	}
}
