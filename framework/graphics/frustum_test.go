package graphics

import (
	"testing"

	"github.com/go-gl/mathgl/mgl32"
)

func TestIsCuboidInFrustum_CenterCalculation(t *testing.T) {
	// Create a simple frustum that accepts anything in a reasonable range
	// We'll create a camera looking down -Z with a wide FOV
	camera := NewCamera(mgl32.DegToRad(90), 1.0)
	camera.Transform().Translation = mgl32.Vec3{0, 0, 0}

	frustum := FrustumFromCamera(camera)

	tests := []struct {
		name     string
		mins     mgl32.Vec3
		maxs     mgl32.Vec3
		expected bool
		reason   string
	}{
		{
			name:     "centered box in front of camera",
			mins:     mgl32.Vec3{-5, -5, -10},
			maxs:     mgl32.Vec3{5, 5, -5},
			expected: true,
			reason:   "box directly in front should be visible",
		},
		{
			name:     "box at origin",
			mins:     mgl32.Vec3{-1, -1, -1},
			maxs:     mgl32.Vec3{1, 1, 1},
			expected: false,
			reason:   "box at camera origin is clipped by near plane (z=-0.2)",
		},
		{
			name:     "box behind camera",
			mins:     mgl32.Vec3{-5, -5, 5},
			maxs:     mgl32.Vec3{5, 5, 10},
			expected: false,
			reason:   "box completely behind camera should not be visible",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := frustum.IsCuboidInFrustum(tt.mins, tt.maxs)
			if result != tt.expected {
				t.Errorf("IsCuboidInFrustum() = %v, expected %v: %s", result, tt.expected, tt.reason)

				// Debug info
				center := tt.mins.Add(tt.maxs).Mul(0.5)
				t.Logf("  Bounding box: mins=%v, maxs=%v", tt.mins, tt.maxs)
				t.Logf("  Calculated center: %v", center)
			}
		})
	}
}

func TestIsCuboidInFrustum_EdgeCases(t *testing.T) {
	camera := NewCamera(mgl32.DegToRad(90), 16.0/9.0)
	camera.Transform().Translation = mgl32.Vec3{0, 0, 0}

	frustum := FrustumFromCamera(camera)

	tests := []struct {
		name     string
		mins     mgl32.Vec3
		maxs     mgl32.Vec3
		expected bool
	}{
		{
			name:     "very far box in front",
			mins:     mgl32.Vec3{-100, -100, -1000},
			maxs:     mgl32.Vec3{100, 100, -900},
			expected: true,
		},
		{
			name:     "box at left edge of frustum",
			mins:     mgl32.Vec3{-50, -5, -60},
			maxs:     mgl32.Vec3{-40, 5, -50},
			expected: true,
		},
		{
			name:     "box at right edge of frustum",
			mins:     mgl32.Vec3{40, -5, -60},
			maxs:     mgl32.Vec3{50, 5, -50},
			expected: true,
		},
		{
			name:     "box at top edge of frustum",
			mins:     mgl32.Vec3{-5, 40, -60},
			maxs:     mgl32.Vec3{5, 50, -50},
			expected: true,
		},
		{
			name:     "box at bottom edge of frustum",
			mins:     mgl32.Vec3{-5, -50, -60},
			maxs:     mgl32.Vec3{5, -40, -50},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := frustum.IsCuboidInFrustum(tt.mins, tt.maxs)
			if result != tt.expected {
				center := tt.mins.Add(tt.maxs).Mul(0.5)
				t.Errorf("IsCuboidInFrustum() = %v, expected %v", result, tt.expected)
				t.Logf("  Bounding box: mins=%v, maxs=%v, center=%v", tt.mins, tt.maxs, center)
			}
		})
	}
}

func TestIsPointInFrustum(t *testing.T) {
	camera := NewCamera(mgl32.DegToRad(90), 1.0)
	camera.Transform().Translation = mgl32.Vec3{0, 0, 0}

	frustum := FrustumFromCamera(camera)

	tests := []struct {
		name     string
		point    mgl32.Vec3
		expected bool
	}{
		{
			name:     "point in front of camera",
			point:    mgl32.Vec3{0, 0, -10},
			expected: true,
		},
		{
			name:     "point at camera origin",
			point:    mgl32.Vec3{0, 0, 0},
			expected: false,
		},
		{
			name:     "point behind camera",
			point:    mgl32.Vec3{0, 0, 10},
			expected: false,
		},
		{
			name:     "point to the left",
			point:    mgl32.Vec3{-5, 0, -10},
			expected: true,
		},
		{
			name:     "point to the right",
			point:    mgl32.Vec3{5, 0, -10},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := frustum.IsPointInFrustum(tt.point[0], tt.point[1], tt.point[2])
			if result != tt.expected {
				t.Errorf("IsPointInFrustum(%v) = %v, expected %v", tt.point, result, tt.expected)
			}
		})
	}
}
