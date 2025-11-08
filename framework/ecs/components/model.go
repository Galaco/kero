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

	// Phase 2: Direct mesh reference (replaces bridge lookup)
	// Using interface{} to avoid mesh package dependency in components.
	// Renderer casts this to *mesh.ModelInstance when needed.
	ModelInstanceHandle interface{} // Mesh model instance (nil = not initialized)
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

// ============================================================================
// ModelInstance Handle Management
// ============================================================================

// SetModelInstance stores the mesh model instance handle for this entity.
// This should be called during entity creation.
func (m *Model) SetModelInstance(instance interface{}) {
	m.ModelInstanceHandle = instance
}

// GetModelInstance retrieves the mesh model instance handle.
// Returns nil if no model instance is attached.
// Renderer should cast this to *mesh.ModelInstance.
func (m *Model) GetModelInstance() interface{} {
	return m.ModelInstanceHandle
}

// HasModelInstance checks if a model instance is attached to this component.
// Returns true if the handle is non-nil.
func (m *Model) HasModelInstance() bool {
	return m.ModelInstanceHandle != nil
}
