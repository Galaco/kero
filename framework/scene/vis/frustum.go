package vis

import (
	"github.com/galaco/kero/framework/graphics"
	"github.com/go-gl/mathgl/mgl32"
	"math"
)

const (
	planeRight    = 0
	planeLeft     = 1
	planeBottom   = 2
	planeTop      = 3
	planeBack     = 4
	planeFront    = 5
	planeNormalX  = 0
	planeNormalY  = 1
	planeNormalZ  = 2
	planeToOrigin = 3
)

// Frustum
// based on https://gist.github.com/jimmikaelkael/2e4ffa5712d61816c7ca
type Frustum struct {
	planes [6][4]float32
}

// IsCuboidInFrustum
func (frustum *Frustum) IsLeafInFrustum(mins, maxs mgl32.Vec3) bool {
	return frustum.IsCuboidInFrustum(mins, maxs)
}

// IsCuboidInFrustum
// NOTE: This fails for leafs where all points sit outside the frustum but the contents is actually inside it
func (frustum *Frustum) IsCuboidInFrustum(mins, maxs mgl32.Vec3) bool {
	center := maxs.Sub(mins)
	if frustum.IsPointInFrustum(center[0], center[1], center[2]) {
		return true
	}

	if frustum.IsPointInFrustum(mins[0], mins[1], mins[2]) {
		return true
	}
	if frustum.IsPointInFrustum(maxs[0], mins[1], mins[2]) {
		return true
	}
	if frustum.IsPointInFrustum(mins[0], maxs[1], mins[2]) {
		return true
	}
	if frustum.IsPointInFrustum(maxs[0], maxs[1], mins[2]) {
		return true
	}
	if frustum.IsPointInFrustum(mins[0], mins[1], maxs[2]) {
		return true
	}
	if frustum.IsPointInFrustum(maxs[0], mins[1], maxs[2]) {
		return true
	}
	if frustum.IsPointInFrustum(mins[0], maxs[1], maxs[2]) {
		return true
	}
	if frustum.IsPointInFrustum(maxs[0], maxs[1], maxs[2]) {
		return true
	}

	return false
}

func (frustum *Frustum) IsPointInFrustum(x, y, z float32) bool {
	// Go through all the sides of the frustum
	for i := 0; i < 6; i++ {
		// Calculate the plane equation and check if the point is behind a side of the frustum
		if frustum.planes[i][planeNormalX]*x+frustum.planes[i][planeNormalY]*y+frustum.planes[i][planeNormalZ]*z+frustum.planes[i][planeToOrigin] <= 0 {
			// The point was behind a side, so it ISN'T in the frustum
			return false
		}
	}

	// The point was inside of the frustum (In front of ALL the sides of the frustum)
	return true
}

func (frustum *Frustum) extractPlanes(modelView mgl32.Mat4, proj mgl32.Mat4) {
	clip := mgl32.Mat4{}

	clip[0] = modelView[0]*proj[0] + modelView[1]*proj[4] + modelView[2]*proj[8] + modelView[3]*proj[12]
	clip[1] = modelView[0]*proj[1] + modelView[1]*proj[5] + modelView[2]*proj[9] + modelView[3]*proj[13]
	clip[2] = modelView[0]*proj[2] + modelView[1]*proj[6] + modelView[2]*proj[10] + modelView[3]*proj[14]
	clip[3] = modelView[0]*proj[3] + modelView[1]*proj[7] + modelView[2]*proj[11] + modelView[3]*proj[15]

	clip[4] = modelView[4]*proj[0] + modelView[5]*proj[4] + modelView[6]*proj[8] + modelView[7]*proj[12]
	clip[5] = modelView[4]*proj[1] + modelView[5]*proj[5] + modelView[6]*proj[9] + modelView[7]*proj[13]
	clip[6] = modelView[4]*proj[2] + modelView[5]*proj[6] + modelView[6]*proj[10] + modelView[7]*proj[14]
	clip[7] = modelView[4]*proj[3] + modelView[5]*proj[7] + modelView[6]*proj[11] + modelView[7]*proj[15]

	clip[8] = modelView[8]*proj[0] + modelView[9]*proj[4] + modelView[10]*proj[8] + modelView[11]*proj[12]
	clip[9] = modelView[8]*proj[1] + modelView[9]*proj[5] + modelView[10]*proj[9] + modelView[11]*proj[13]
	clip[10] = modelView[8]*proj[2] + modelView[9]*proj[6] + modelView[10]*proj[10] + modelView[11]*proj[14]
	clip[11] = modelView[8]*proj[3] + modelView[9]*proj[7] + modelView[10]*proj[11] + modelView[11]*proj[15]

	clip[12] = modelView[12]*proj[0] + modelView[13]*proj[4] + modelView[14]*proj[8] + modelView[15]*proj[12]
	clip[13] = modelView[12]*proj[1] + modelView[13]*proj[5] + modelView[14]*proj[9] + modelView[15]*proj[13]
	clip[14] = modelView[12]*proj[2] + modelView[13]*proj[6] + modelView[14]*proj[10] + modelView[15]*proj[14]
	clip[15] = modelView[12]*proj[3] + modelView[13]*proj[7] + modelView[14]*proj[11] + modelView[15]*proj[15]

	// Now we actually want to get the sides of the frustum.  To do this we take
	// the clipping planes we received above and extract the sides from them.

	// This will extract the RIGHT side of the frustum
	frustum.planes[planeRight][planeNormalX] = clip[3] - clip[0]
	frustum.planes[planeRight][planeNormalY] = clip[7] - clip[4]
	frustum.planes[planeRight][planeNormalZ] = clip[11] - clip[8]
	frustum.planes[planeRight][planeToOrigin] = clip[15] - clip[12]

	// Now that we have a normal (A,B,C) and a distance (D) to the plane,
	// we want to normalize that normal and distance.

	// Normalize the RIGHT side
	frustum.normalizePlane(planeRight)

	// This will extract the LEFT side of the frustum
	frustum.planes[planeLeft][planeNormalX] = clip[3] + clip[0]
	frustum.planes[planeLeft][planeNormalY] = clip[7] + clip[4]
	frustum.planes[planeLeft][planeNormalZ] = clip[11] + clip[8]
	frustum.planes[planeLeft][planeToOrigin] = clip[15] + clip[12]

	// Normalize the LEFT side
	frustum.normalizePlane(planeLeft)

	// This will extract the BOTTOM side of the frustum
	frustum.planes[planeBottom][planeNormalX] = clip[3] + clip[1]
	frustum.planes[planeBottom][planeNormalY] = clip[7] + clip[5]
	frustum.planes[planeBottom][planeNormalZ] = clip[11] + clip[9]
	frustum.planes[planeBottom][planeToOrigin] = clip[15] + clip[13]

	// Normalize the BOTTOM side
	frustum.normalizePlane(planeBottom)

	// This will extract the TOP side of the frustum
	frustum.planes[planeTop][planeNormalX] = clip[3] - clip[1]
	frustum.planes[planeTop][planeNormalY] = clip[7] - clip[5]
	frustum.planes[planeTop][planeNormalZ] = clip[11] - clip[9]
	frustum.planes[planeTop][planeToOrigin] = clip[15] - clip[13]

	// Normalize the TOP side
	frustum.normalizePlane(planeTop)

	// This will extract the BACK side of the frustum
	frustum.planes[planeBack][planeNormalX] = clip[3] - clip[2]
	frustum.planes[planeBack][planeNormalY] = clip[7] - clip[6]
	frustum.planes[planeBack][planeNormalZ] = clip[11] - clip[10]
	frustum.planes[planeBack][planeToOrigin] = clip[15] - clip[14]

	// Normalize the BACK side
	frustum.normalizePlane(planeBack)

	// This will extract the FRONT side of the frustum
	frustum.planes[planeFront][planeNormalX] = clip[3] + clip[2]
	frustum.planes[planeFront][planeNormalY] = clip[7] + clip[6]
	frustum.planes[planeFront][planeNormalZ] = clip[11] + clip[10]
	frustum.planes[planeFront][planeToOrigin] = clip[15] + clip[14]

	// Normalize the FRONT side
	frustum.normalizePlane(planeFront)
}

func (frustum *Frustum) normalizePlane(side int) {
	// Here we calculate the magnitude of the normal to the plane (point A B C)
	// Remember that (A, B, C) is that same thing as the normal's (X, Y, Z).
	// To calculate magnitude you use the equation:  magnitude = sqrt( x^2 + y^2 + z^2)
	magnitude := float32(math.Sqrt(
		float64(frustum.planes[side][planeNormalX]*frustum.planes[side][planeNormalX] +
			frustum.planes[side][planeNormalY]*frustum.planes[side][planeNormalY] +
			frustum.planes[side][planeNormalZ]*frustum.planes[side][planeNormalZ])))
	// Then we divide the plane's values by it's magnitude.
	// This makes it easier to work with.
	frustum.planes[side][planeNormalX] /= magnitude
	frustum.planes[side][planeNormalY] /= magnitude
	frustum.planes[side][planeNormalZ] /= magnitude
	frustum.planes[side][planeToOrigin] /= magnitude
}

// FrustumFromCamera
func FrustumFromCamera(camera *graphics.Camera) *Frustum {
	f := &Frustum{}
	f.extractPlanes(camera.ModelMatrix().Mul4(camera.ViewMatrix()), camera.ProjectionMatrix())

	return f
}
