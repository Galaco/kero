package legacy

import (
	"github.com/galaco/kero/framework/ecs"
	"github.com/galaco/kero/framework/ecs/components"
	"github.com/galaco/kero/framework/entity"
	"github.com/galaco/kero/framework/graphics/mesh"
	"github.com/go-gl/mathgl/mgl32"
)

// Bridge converts between legacy entity.IEntity and ECS entities.
// This enables incremental migration - old and new systems can coexist.
type Bridge struct {
	world *ecs.World

	// Maps ECS Entity ID to legacy entity.IEntity
	ecsToLegacy map[ecs.Entity]entity.IEntity

	// Maps legacy entity pointer to ECS Entity ID
	legacyToECS map[entity.IEntity]ecs.Entity
}

// NewBridge creates a new legacy entity bridge
func NewBridge(world *ecs.World) *Bridge {
	return &Bridge{
		world:       world,
		ecsToLegacy: make(map[ecs.Entity]entity.IEntity),
		legacyToECS: make(map[entity.IEntity]ecs.Entity),
	}
}

// CreateECSEntityFromLegacy creates an ECS entity from a legacy entity.
// This copies the entity's transform, model, and physics properties into ECS components.
func (b *Bridge) CreateECSEntityFromLegacy(legacyEntity entity.IEntity) ecs.Entity {
	// Create ECS entity
	ecsEntity := b.world.CreateEntity()

	// Add Transform component
	transform := components.Transform{
		Position:    legacyEntity.Origin(),
		Orientation: legacyEntity.Angles(),
		Scale:       mgl32.Vec3{1, 1, 1}, // Legacy entities don't expose scale directly
	}
	ecs.AddComponent(b.world, ecsEntity, transform)

	// Add Model component if entity has a model
	if legacyEntity.Model() != nil {
		model := components.Model{
			MeshPath:       legacyEntity.Model().Model.Id, // Use model ID as path
			MaterialID:     0,                             // Will be populated by renderer
			CastShadows:    true,
			ReceiveShadows: true,
			Visible:        true,
			ModelScale:     1.0,
		}
		ecs.AddComponent(b.world, ecsEntity, model)

		// Add Physics component if entity has a rigid body
		if legacyEntity.Model().RigidBody != nil {
			physics := components.Physics{
				Velocity:    mgl32.Vec3{0, 0, 0},  // Bullet manages velocity internally
				Force:       mgl32.Vec3{0, 0, 0},
				AngularVel:  mgl32.Vec3{0, 0, 0},
				Torque:      mgl32.Vec3{0, 0, 0},
				Mass:        legacyEntity.Model().Model.OriginalStudiomodel.Mdl.Header.Mass,
				Restitution: 0.5,
				Friction:    0.5,
				Drag:        0.1,
				AngularDrag: 0.1,
				RigidBodyID: 0, // Note: RigidBody accessed through legacy entity during migration
				IsKinematic: false,
				UseGravity:  true,
				IsSleeping:  false,
				IsStatic:    false,
			}
			ecs.AddComponent(b.world, ecsEntity, physics)
		}
	}

	// Store bidirectional mapping
	b.ecsToLegacy[ecsEntity] = legacyEntity
	b.legacyToECS[legacyEntity] = ecsEntity

	return ecsEntity
}

// GetECSEntity returns the ECS entity ID for a legacy entity
func (b *Bridge) GetECSEntity(legacyEntity entity.IEntity) (ecs.Entity, bool) {
	ecsEntity, exists := b.legacyToECS[legacyEntity]
	return ecsEntity, exists
}

// GetLegacyEntity returns the legacy entity for an ECS entity ID
func (b *Bridge) GetLegacyEntity(ecsEntity ecs.Entity) (entity.IEntity, bool) {
	legacyEntity, exists := b.ecsToLegacy[ecsEntity]
	return legacyEntity, exists
}

// SyncLegacyToECS updates ECS components from legacy entity state.
// Call this after legacy systems modify entity properties.
func (b *Bridge) SyncLegacyToECS(legacyEntity entity.IEntity) {
	ecsEntity, exists := b.legacyToECS[legacyEntity]
	if !exists {
		return
	}

	// Update Transform component
	if transform, ok := ecs.GetComponent[components.Transform](b.world, ecsEntity); ok {
		transform.Position = legacyEntity.Origin()
		transform.Orientation = legacyEntity.Angles()
		// Note: We're modifying the component directly, which is already stored in the ECS
	}
}

// SyncECSToLegacy updates legacy entity state from ECS components.
// Call this after ECS systems modify components.
func (b *Bridge) SyncECSToLegacy(ecsEntity ecs.Entity) {
	legacyEntity, exists := b.ecsToLegacy[ecsEntity]
	if !exists {
		return
	}

	// Update legacy entity transform from ECS component
	if transform, ok := ecs.GetComponent[components.Transform](b.world, ecsEntity); ok {
		legacyTransform := legacyEntity.Transform()
		legacyTransform.Translation = transform.Position
		legacyTransform.Orientation = transform.Orientation
	}
}

// SyncAllECSToLegacy syncs all ECS entities back to their legacy counterparts.
// Useful for systems that still read from legacy entities.
func (b *Bridge) SyncAllECSToLegacy() {
	for ecsEntity := range b.ecsToLegacy {
		b.SyncECSToLegacy(ecsEntity)
	}
}

// SyncAllLegacyToECS syncs all legacy entities to their ECS counterparts.
// Useful after legacy systems have updated entity state.
func (b *Bridge) SyncAllLegacyToECS() {
	for legacyEntity := range b.legacyToECS {
		b.SyncLegacyToECS(legacyEntity)
	}
}

// Clear removes all mappings
func (b *Bridge) Clear() {
	b.ecsToLegacy = make(map[ecs.Entity]entity.IEntity)
	b.legacyToECS = make(map[entity.IEntity]ecs.Entity)
}

// EntityCount returns the number of bridged entities
func (b *Bridge) EntityCount() int {
	return len(b.ecsToLegacy)
}

// ConvertModelInstanceToComponents extracts component data from a legacy ModelInstance.
// This is a helper for creating ECS entities without full legacy entities.
func ConvertModelInstanceToComponents(modelInstance *mesh.ModelInstance, initialTransform mgl32.Mat4) (components.Model, components.Physics, bool) {
	model := components.Model{
		MeshPath:       modelInstance.Model.Id,
		MaterialID:     0,
		CastShadows:    true,
		ReceiveShadows: true,
		Visible:        true,
		ModelScale:     1.0,
	}

	hasPhysics := modelInstance.RigidBody != nil
	var physics components.Physics

	if hasPhysics {
		mass := modelInstance.Model.OriginalStudiomodel.Mdl.Header.Mass
		physics = components.Physics{
			Velocity:    mgl32.Vec3{0, 0, 0},
			Force:       mgl32.Vec3{0, 0, 0},
			AngularVel:  mgl32.Vec3{0, 0, 0},
			Torque:      mgl32.Vec3{0, 0, 0},
			Mass:        mass,
			Restitution: 0.5,
			Friction:    0.5,
			Drag:        0.1,
			AngularDrag: 0.1,
			RigidBodyID: 0, // Note: RigidBody accessed through legacy entity during migration
			IsKinematic: false,
			UseGravity:  true,
			IsSleeping:  false,
			IsStatic:    mass == 0,
		}
	}

	return model, physics, hasPhysics
}
