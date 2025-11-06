package components

import "github.com/galaco/kero/framework/ecs"

// Model component for visual representation.
// Entities with this component are rendered to the screen.
type Model struct {
	// Resource references
	MeshPath   string // Path to mesh file
	MaterialID uint32 // Material identifier

	// Rendering properties
	CastShadows    bool
	ReceiveShadows bool
	RenderLayer    uint8  // For sorting/filtering (0 = default, higher = later)
	Visible        bool

	// GPU resource handle (populated by renderer)
	GPUMeshID     uint32 // OpenGL VAO/buffer ID
	GPUMaterialID uint32 // OpenGL texture/shader ID

	// Model specific data
	ModelScale float32 // Additional scale factor (beyond Transform.Scale)
}

// IsComponent implements the ecs.Component marker interface
func (Model) IsComponent() {}

// Register Model component with the ECS system
func init() {
	ecs.RegisterComponent[Model](ecs.ComponentTypeModel)
}

// NewModel creates a Model component with default values
func NewModel(meshPath string) Model {
	return Model{
		MeshPath:       meshPath,
		MaterialID:     0,
		CastShadows:    true,
		ReceiveShadows: true,
		RenderLayer:    0,
		Visible:        true,
		GPUMeshID:      0,
		GPUMaterialID:  0,
		ModelScale:     1.0,
	}
}

// IsLoaded returns true if GPU resources are loaded
func (m *Model) IsLoaded() bool {
	return m.GPUMeshID != 0
}
