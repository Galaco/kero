package physics

import (
	"github.com/galaco/kero/framework/physics/collision/bullet"
	"github.com/galaco/kero/framework/scene"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	surfNoDraw = 0x0080
	surfTrigger = 0x0040
	surfSkip = 0x0200
)

type BspCollisionMesh struct {
	indices [][]bullet.BulletPhysicsIndice
	vertices [][]bullet.BulletVec3
	childShapeHandles []bullet.BulletCollisionShapeHandle
	RigidBodyHandle bullet.BulletRigidBodyHandle
}

func generateBspCollisionMesh(scene *scene.StaticScene) *BspCollisionMesh {
	var verts []float32

	indices := make([][]bullet.BulletPhysicsIndice, len(scene.ClusterLeafs))
	vertices := make([][]bullet.BulletVec3, len(scene.ClusterLeafs))
	childShapeHandles := make([]bullet.BulletCollisionShapeHandle, len(scene.ClusterLeafs))

	bspCompoundShape := bullet.BulletNewCompoundShape()

	for idx,l := range scene.ClusterLeafs {
		index := int64(0)
		indices[idx] = make([]bullet.BulletPhysicsIndice, 0)
		vertices[idx] = make([]bullet.BulletVec3, 0)
		for _,f := range l.Faces {
			// Much more optimisation can be done here to discard non-solid faces
			if scene.RawBsp.TexInfos()[f.RawFace().TexInfo].Flags & surfNoDraw != 0 {
				continue
			}

			verts = scene.RawBsp.Mesh().Vertices()[f.Offset() : f.Offset()+f.Length()]
			if len(verts) < 9 {
				// Something is broken here; how can a triangle not have 3 verts?
				continue
			}
			vertices[idx] = append(vertices[idx],
				bullet.Vec3ToBullet(mgl32.Vec3{
					verts[0],
					verts[1],
					verts[2],
				}),
				bullet.Vec3ToBullet(mgl32.Vec3{
					verts[3],
					verts[4],
					verts[5],
				}),
				bullet.Vec3ToBullet(mgl32.Vec3{
					verts[6],
					verts[7],
					verts[8],
				}),
			)
			indices[idx] = append(indices[idx], bullet.BulletPhysicsIndice(index), bullet.BulletPhysicsIndice(index + 1), bullet.BulletPhysicsIndice(index + 2))
			index += 3
		}
		if len(vertices[idx]) < 9 {
			continue
		}

		childShapeHandles[idx] = bullet.BulletNewStaticTriangleShape(indices[idx], vertices[idx], int64(len(indices[idx]) / 3), int64(len(vertices[idx])))

		bullet.BulletAddChildToCompoundShape(bspCompoundShape, childShapeHandles[idx], mgl32.Vec3{0,0,0}, mgl32.QuatIdent())
	}

	return &BspCollisionMesh{
		vertices: vertices,
		indices: indices,
		childShapeHandles: childShapeHandles,
		RigidBodyHandle: bullet.NewRigidBody(0, bspCompoundShape),
	}
}