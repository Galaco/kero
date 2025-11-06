package mesh

import (
	"github.com/galaco/kero/framework/physics/collision"
	"github.com/galaco/studiomodel"
	"github.com/go-gl/mathgl/mgl32"
	"math"
)

type ModelInstance struct {
	Model     *Model
	RigidBody collision.RigidBody
}

type Model struct {
	Id                  string
	OriginalStudiomodel *studiomodel.StudioModel
	meshes              []*BasicMesh
	materials           []string
	rigidBody           collision.RigidBody
	boundsMins          mgl32.Vec3
	boundsMaxs          mgl32.Vec3
	boundsComputed      bool
}

func (model *Model) Meshes() []*BasicMesh {
	return model.meshes
}

func (model *Model) Materials() []string {
	return model.materials
}

func (model *Model) AddMesh(m *BasicMesh) {
	model.meshes = append(model.meshes, m)
}

func (model *Model) AddMaterial(m string) {
	model.materials = append(model.materials, m)
}

func (model *Model) RigidBody() collision.RigidBody {
	return model.rigidBody
}

func (model *Model) AddRigidBody(body collision.RigidBody) {
	model.rigidBody = body
}

// ComputeBounds calculates the axis-aligned bounding box for all meshes in the model
func (model *Model) ComputeBounds() {
	if model.boundsComputed || len(model.meshes) == 0 {
		return
	}

	// Initialize with extreme values
	mins := mgl32.Vec3{math.MaxFloat32, math.MaxFloat32, math.MaxFloat32}
	maxs := mgl32.Vec3{-math.MaxFloat32, -math.MaxFloat32, -math.MaxFloat32}

	// Iterate through all meshes and vertices
	for _, mesh := range model.meshes {
		verts := mesh.Vertices()
		for i := 0; i < len(verts); i += 3 {
			x, y, z := verts[i], verts[i+1], verts[i+2]

			// Update mins
			if x < mins[0] {
				mins[0] = x
			}
			if y < mins[1] {
				mins[1] = y
			}
			if z < mins[2] {
				mins[2] = z
			}

			// Update maxs
			if x > maxs[0] {
				maxs[0] = x
			}
			if y > maxs[1] {
				maxs[1] = y
			}
			if z > maxs[2] {
				maxs[2] = z
			}
		}
	}

	model.boundsMins = mins
	model.boundsMaxs = maxs
	model.boundsComputed = true
}

// Bounds returns the computed bounding box for this model
// Returns mins, maxs. If bounds haven't been computed, computes them first.
func (model *Model) Bounds() (mgl32.Vec3, mgl32.Vec3) {
	if !model.boundsComputed {
		model.ComputeBounds()
	}
	return model.boundsMins, model.boundsMaxs
}

func NewModel(id string, originalStudioModel *studiomodel.StudioModel) *Model {
	return &Model{
		Id:                  id,
		OriginalStudiomodel: originalStudioModel,
	}
}
