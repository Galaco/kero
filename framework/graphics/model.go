package graphics

type Model struct {
	Id        string
	meshes    []*Mesh
	materials []string
}

func (model *Model) Meshes() []*Mesh {
	return model.meshes
}

func (model *Model) Materials() []string {
	return model.materials
}

func (model *Model) AddMesh(m *Mesh) {
	model.meshes = append(model.meshes, m)
}

func (model *Model) AddMaterial(m string) {
	model.materials = append(model.materials, m)
}

func NewModel(id string) *Model {
	return &Model{
		Id: id,
	}
}
