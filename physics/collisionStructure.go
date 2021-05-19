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

func generateBspCollisionMesh(scene *scene.StaticScene) bullet.BulletRigidBodyHandle {
	var verts []float32

	indices := make([]bullet.BulletPhysicsIndice, 0)
	vertices := make([]bullet.BulletVec3, 0)

	bspCompoundShape := bullet.BulletNewCompoundShape()

	for _,l := range scene.ClusterLeafs {
		index := int64(0)
		vertices = make([]bullet.BulletVec3, 0)
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
			vertices = append(vertices,
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
			indices = append(indices, bullet.BulletPhysicsIndice(index), bullet.BulletPhysicsIndice(index + 1), bullet.BulletPhysicsIndice(index + 2))
			index += 3
		}
		if len(vertices) < 9 {
			continue
		}

		bullet.BulletAddChildToCompoundShape(bspCompoundShape, bullet.BulletNewStaticTriangleShape(indices, vertices, int64(len(indices) / 3), int64(len(vertices))), mgl32.Vec3{0,0,0}, mgl32.QuatIdent())

	}

	bullet.BulletAddChildToCompoundShape(bspCompoundShape, bullet.BulletNewStaticPlaneShape(mgl32.Vec3{1,-10,1}, 5), mgl32.Vec3{0,0,0}, mgl32.QuatIdent())

	return bullet.NewRigidBody(0, bspCompoundShape)
}