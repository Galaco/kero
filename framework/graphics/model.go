package graphics

type Model struct {
	meshes []*Mesh
	materials []string
}

func (model *Model) AddMesh(m *Mesh) {
	model.meshes = append(model.meshes, m)
}

func (model *Model) AddMaterial(m string) {
	model.materials = append(model.materials, m)
}

func NewModel(id string) *Model{
	return &Model{}
}
