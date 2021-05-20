package physics

import (
	"github.com/galaco/bsp"
	"github.com/galaco/bsp/lumps"
	"github.com/galaco/bsp/primitives/brush"
	"github.com/galaco/bsp/primitives/plane"
	"github.com/galaco/kero/framework/physics/collision/bullet"
	"github.com/galaco/kero/framework/scene"
	"github.com/go-gl/mathgl/mgl32"
	"math"
	"sync"
)

const (
	surfNoDraw = 0x0080
	surfTrigger = 0x0040
	surfSkip = 0x0200
)

type BspCollisionMesh struct {
	indices [][]bullet.BulletPhysicsIndice
	vertices [][]mgl32.Vec3
	childShapeHandles []bullet.BulletCollisionShapeHandle
	// RigidBodyHandle bullet.BulletRigidBodyHandle
	RigidBodyHandles []bullet.BulletRigidBodyHandle
}

func generateBspCollisionMesh(scene *scene.StaticScene) *BspCollisionMesh {
	brushes := scene.RawBsp.File().Lump(bsp.LumpBrushes).(*lumps.Brush).GetData()
	brushSides := scene.RawBsp.File().Lump(bsp.LumpBrushSides).(*lumps.BrushSide).GetData()
	planes := scene.RawBsp.File().Lump(bsp.LumpPlanes).(*lumps.Planes).GetData()

	childShapeHandles := make([]bullet.BulletCollisionShapeHandle, 0)
	handles := make([]bullet.BulletRigidBodyHandle, 0)

	wg := sync.WaitGroup{}

	verts := make([][]mgl32.Vec3, len(brushes))
	wg.Add(len(brushes))

	asyncVertsFromPlanes := func (b brush.Brush, idx int) {
		if b.Contents & 0x1 <= 0 || b.NumSides < 1 {
			wg.Done()
			return
		}
		sides := brushSides[b.FirstSide:b.FirstSide + b.NumSides]
		planeNormals := make([]plane.Plane, len(sides))

		for i,side := range sides {
			planeNormals[i] = planes[side.PlaneNum]
		}

		verts[idx] = getVerticesFromPlaneEquations(planeNormals)

		wg.Done()
	}

	for idx,b := range brushes {
		go asyncVertsFromPlanes(b, idx)
	}

	wg.Wait()
	for idx := range brushes {
		if verts[idx] == nil || len(verts[idx]) == 0 {
			continue
		}

		h := bullet.BulletNewBrushShape(verts[idx])

		childShapeHandles = append(childShapeHandles, h)
		handles = append(handles, bullet.NewRigidBody(0, h))
	}

	return &BspCollisionMesh{
		vertices: verts,
		childShapeHandles: childShapeHandles,
		RigidBodyHandles: handles,
	}
}


func isPointInsidePlanes(planeEquations []plane.Plane, point mgl32.Vec3, margin float32) bool {
	for i := 0; i < len(planeEquations); i++ {
		dist := (planeEquations[i].Normal.Mul(planeEquations[i].Distance).Dot(point) + 0) - margin
		// flipped operator?!
		if dist > 0. {
			return false
		}
	}
	return true
}

func getVerticesFromPlaneEquations(planeEquations []plane.Plane) (verticesOut []mgl32.Vec3) {
	// brute force:
	for i := 0; i < len(planeEquations); i++ {
		N1 := planeEquations[i].Normal
		D1 := planeEquations[i].Distance

		for j := i + 1; j < len(planeEquations); j++ {
			N2 := planeEquations[j].Normal
			D2 := planeEquations[j].Distance

			for k := j + 1; k < len(planeEquations); k++ {
				N3 := planeEquations[k].Normal
				D3 := planeEquations[k].Distance

				n2n3 := N2.Cross(N3)
				n3n1 := N3.Cross(N1)
				n1n2 := N1.Cross(N2)

				if (n2n3.LenSqr() > 0.0001) && (n3n1.LenSqr() > 0.0001) && (n1n2.LenSqr() > 0.0001) {
					//point P out of 3 plane equations:

					//	d1 ( N2 * N3 ) + d2 ( N3 * N1 ) + d3 ( N1 * N2 )
					//P =  -------------------------------------------------------------------------
					//   N1 . ( N2 * N3 )

					quotient := N1.Dot(n2n3)
					if math.Abs(float64(quotient)) > 0.000001 {
						quotient = -1. / quotient
						//n2n3 = n2n3.Mul(N1[3])
						//n3n1 = n3n1.Mul(N2[3])
						//n1n2 = n1n2.Mul(N3[3])
						n2n3 = n2n3.Mul(D1)
						n3n1 = n3n1.Mul(D2)
						n1n2 = n1n2.Mul(D3)
						potentialVertex := n2n3.Add(n3n1).Add(n1n2).Mul(quotient)

						//check if inside, and replace supportingVertexOut if needed
						if isPointInsidePlanes(planeEquations, potentialVertex, 0.01) {
							verticesOut = append(verticesOut, mgl32.Vec3{-potentialVertex[0], -potentialVertex[1], -potentialVertex[2]})
						}
					}
				}
			}
		}
	}

	return verticesOut
}