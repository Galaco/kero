package physics

import (
	"github.com/galaco/bsp"
	"github.com/galaco/bsp/lumps"
	"github.com/galaco/bsp/primitives/brush"
	"github.com/galaco/bsp/primitives/plane"
	"github.com/galaco/kero/framework/physics/collision/bullet"
	"github.com/galaco/kero/framework/scene"
	"github.com/galaco/studiomodel/mdl"
	"github.com/galaco/studiomodel/phy"
	"github.com/go-gl/mathgl/mgl32"
	"math"
	"sync"
)

type bspCollisionMesh struct {
	indices [][]bullet.BulletPhysicsIndice
	vertices [][]mgl32.Vec3
	childShapeHandles []bullet.BulletCollisionShapeHandle
	RigidBodyHandles []bullet.BulletRigidBodyHandle
}

func generateBspCollisionMesh(scene *scene.StaticScene) *bspCollisionMesh {
	brushes := scene.RawBsp.File().Lump(bsp.LumpBrushes).(*lumps.Brush).GetData()
	brushSides := scene.RawBsp.File().Lump(bsp.LumpBrushSides).(*lumps.BrushSide).GetData()
	planes := scene.RawBsp.File().Lump(bsp.LumpPlanes).(*lumps.Planes).GetData()

	childShapeHandles := make([]bullet.BulletCollisionShapeHandle, 0)
	handles := make([]bullet.BulletRigidBodyHandle, 0)

	wg := sync.WaitGroup{}

	verts := make([][]mgl32.Vec3, len(brushes))
	wg.Add(len(brushes))

	asyncVertsFromPlanes := func (b brush.Brush, idx int) {
		if b.Contents & bsp.CONTENTS_SOLID <= 0 || b.NumSides < 1 {
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

	// This loop *could* be done inside asyncVertsFromPlanes, but maybe these bullet CGo calls aren't threadsafe
	for idx := range brushes {
		if verts[idx] == nil || len(verts[idx]) == 0 {
			continue
		}

		h := bullet.BulletNewBrushShape(verts[idx])

		childShapeHandles = append(childShapeHandles, h)
		handles = append(handles, bullet.NewRigidBody(0, h))
	}

	return &bspCollisionMesh{
		vertices: verts,
		childShapeHandles: childShapeHandles,
		RigidBodyHandles: handles,
	}
}

type studiomodelCollisionMesh struct {
	vertices            [][]mgl32.Vec3
	compountShapeHandle bullet.BulletCollisionShapeHandle
}

func generateCollisionMeshFromStudiomodelPhy(phy *phy.Phy) studiomodelCollisionMesh {
	parts := make([]bullet.BulletCollisionShapeHandle, 0)

	faceOffset := int32(0)
	verts := make([][]mgl32.Vec3, len(phy.TriangleFaceHeaders))
	for idx,header := range phy.TriangleFaceHeaders {
		verts[idx] = make([]mgl32.Vec3, 0)
		for _,face := range phy.TriangleFaces[faceOffset:faceOffset + header.FaceCount] {
			//  PHY vertices use a different scaling space!!!
			verts[idx] = append(verts[idx],
				transformPhyVertex(nil, mgl32.Vec3{
					phy.Vertices[int32(face.V1)][0],
					phy.Vertices[int32(face.V1)][1],
					phy.Vertices[int32(face.V1)][2],
				}),
				transformPhyVertex(nil, mgl32.Vec3{
					phy.Vertices[int32(face.V2)][0],
					phy.Vertices[int32(face.V2)][1],
					phy.Vertices[int32(face.V2)][2],
				}),
				transformPhyVertex(nil, mgl32.Vec3{
					phy.Vertices[int32(face.V3)][0],
					phy.Vertices[int32(face.V3)][1],
					phy.Vertices[int32(face.V3)][2],
				}),
			)
		}
		faceOffset += header.FaceCount

		part := bullet.BulletNewConvexHullShape()
		part.AddVertices(verts[idx])
	}

	mesh := studiomodelCollisionMesh {
		vertices:            verts,
		compountShapeHandle: bullet.BulletNewCompoundShape(),
	}

	shape := bullet.BulletNewCompoundShape()
	for _,i := range parts {
		bullet.BulletAddChildToCompoundShape(shape, i, mgl32.Vec3{}, mgl32.Quat{})
	}

	return mesh
}

func transformPhyVertex(bone *mdl.Bone, vertex mgl32.Vec3) (out mgl32.Vec3) {
	out[0] = 1 / 0.0254 * vertex[0]
	out[1] = 1 / 0.0254 * vertex[2]
	out[2] = 1 / 0.0254 * -vertex[1]
	if bone != nil {
		out = vectorITransform(out, bone.PoseToBone)
	} else {
		out[0] = 1 / 0.0254 * vertex[2]
		out[1] = 1 / 0.0254 * -vertex[0]
		out[2] = 1 / 0.0254 * -vertex[1]
	}
	return out
}

func vectorITransform (in1 mgl32.Vec3, in2 mgl32.Mat3x4) (out mgl32.Vec3) {
	t := mgl32.Vec3{}
	t[0] = in1[0] - in2.Col(3)[0]
	t[1] = in1[1] - in2.Col(3)[1]
	t[2] = in1[2] - in2.Col(3)[2]

	out[0] = t[0] * in2.Col(0)[0] + t[1] * in2.Col(0)[1] + t[2] * in2.Col(0)[2]
	out[1] = t[0] * in2.Col(1)[0] + t[1] * in2.Col(1)[1] + t[2] * in2.Col(1)[2]
	out[2] = t[0] * in2.Col(2)[0] + t[1] * in2.Col(2)[1] + t[2] * in2.Col(2)[2]

	return out
}


func isPointInsidePlanes(planeEquations []plane.Plane, point mgl32.Vec3, margin float32) bool {
	for i := 0; i < len(planeEquations); i++ {
		dist := (planeEquations[i].Normal.Mul(-planeEquations[i].Distance).Dot(point) + 0) - margin
		if dist > 0. {
			return false
		}
	}
	return true
}

func getVerticesFromPlaneEquations(planeEquations []plane.Plane) (verticesOut []mgl32.Vec3) {
	// brute force:
	var N1, N2, N3 mgl32.Vec3
	var D1, D2, D3 float32
	for i := 0; i < len(planeEquations); i++ {
		N1 = planeEquations[i].Normal
		D1 = -planeEquations[i].Distance

		for j := i + 1; j < len(planeEquations); j++ {
			N2 = planeEquations[j].Normal
			D2 = -planeEquations[j].Distance

			for k := j + 1; k < len(planeEquations); k++ {
				N3 = planeEquations[k].Normal
				D3 = -planeEquations[k].Distance

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
						n2n3 = n2n3.Mul(D1)
						n3n1 = n3n1.Mul(D2)
						n1n2 = n1n2.Mul(D3)
						potentialVertex := n2n3.Add(n3n1).Add(n1n2).Mul(quotient)

						//check if inside, and replace supportingVertexOut if needed
						if isPointInsidePlanes(planeEquations, potentialVertex, 0.01) {
							verticesOut = append(verticesOut, potentialVertex)
						}
					}
				}
			}
		}
	}

	return verticesOut
}