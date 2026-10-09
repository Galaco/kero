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

// SubMesh is a range of a model's indices that is drawn with one material
type SubMesh struct {
	IndexOffset int
	IndexCount  int
}

type Model struct {
	Id                  string
	OriginalStudiomodel *studiomodel.StudioModel
	mesh                *BasicMesh
	subMeshes           []SubMesh
	materials           [][]string // the material of each sub-mesh, for each skin
	rigidBody           collision.RigidBody
	boundsMins          mgl32.Vec3
	boundsMaxs          mgl32.Vec3
	boundsComputed      bool
}

// Mesh returns the vertices shared by all sub-meshes, and the indices of every sub-mesh one after another
func (model *Model) Mesh() *BasicMesh {
	return model.mesh
}

// SubMeshes returns the ranges of Mesh's indices that are each drawn with one material
func (model *Model) SubMeshes() []SubMesh {
	return model.subMeshes
}

// NumSkins returns how many skins the model has
func (model *Model) NumSkins() int {
	return len(model.materials)
}

// Skin returns skin if the model has it, otherwise the default skin 0, as the engine does
func (model *Model) Skin(skin int) int {
	if skin < 0 || skin >= len(model.materials) {
		return 0
	}
	return skin
}

// Materials returns the material of each sub-mesh with a skin
func (model *Model) Materials(skin int) []string {
	if len(model.materials) == 0 {
		return nil
	}
	return model.materials[model.Skin(skin)]
}

// AddSubMesh adds indices into Mesh's vertices that are drawn with a material, given for each skin.
// Every sub-mesh must be given the same number of skins.
func (model *Model) AddSubMesh(indices []uint32, skinMaterials []string) {
	model.subMeshes = append(model.subMeshes, SubMesh{
		IndexOffset: len(model.mesh.Indices()),
		IndexCount:  len(indices),
	})
	model.mesh.AddIndice(indices...)
	for skin, material := range skinMaterials {
		if skin == len(model.materials) {
			model.materials = append(model.materials, nil)
		}
		model.materials[skin] = append(model.materials[skin], material)
	}
}

func (model *Model) RigidBody() collision.RigidBody {
	return model.rigidBody
}

func (model *Model) AddRigidBody(body collision.RigidBody) {
	model.rigidBody = body
}

// ComputeBounds calculates the axis-aligned bounding box of the model's vertices
func (model *Model) ComputeBounds() {
	verts := model.mesh.Vertices()
	if model.boundsComputed || len(verts) == 0 {
		return
	}

	// Initialize with extreme values
	mins := mgl32.Vec3{math.MaxFloat32, math.MaxFloat32, math.MaxFloat32}
	maxs := mgl32.Vec3{-math.MaxFloat32, -math.MaxFloat32, -math.MaxFloat32}

	for i := 0; i < len(verts); i += 3 {
		for axis := 0; axis < 3; axis++ {
			mins[axis] = min(mins[axis], verts[i+axis])
			maxs[axis] = max(maxs[axis], verts[i+axis])
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
		mesh:                NewMesh(),
	}
}
