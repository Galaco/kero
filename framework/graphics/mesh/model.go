package mesh

import (
	"github.com/galaco/kero/framework/physics/collision"
	"github.com/galaco/studiomodel"
)

type ModelInstance struct {
	Model *Model
	RigidBody collision.RigidBody
}

type Model struct {
	Id        string
	OriginalStudiomodel *studiomodel.StudioModel
	meshes    []*BasicMesh
	materials []string
	rigidBody collision.RigidBody
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

func NewModel(id string, originalStudioModel *studiomodel.StudioModel) *Model {
	return &Model{
		Id: id,
		OriginalStudiomodel: originalStudioModel,
	}
}
