package graphics

import (
	"testing"

	"github.com/go-gl/mathgl/mgl32"
)

func TestCamera_ModelMatrix(t *testing.T) {
	c := NewCamera(90, 16/9)
	actual := c.ModelMatrix()
	expected := mgl32.Ident4()

	if !actual.ApproxEqual(expected) {
		t.Error("unexpected matrix from camera modelMatrix, expected an identity matrix")
	}
}

func TestCamera_ViewMatrix(t *testing.T) {
	t.Skip()
}

func TestCamera_ProjectionMatrix(t *testing.T) {
	t.Skip()
}

func TestCamera_Update(t *testing.T) {
	t.Skip()
}
