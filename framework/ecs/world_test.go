package ecs_test

import (
	"testing"

	"github.com/galaco/kero/framework/ecs"
	"github.com/galaco/kero/framework/ecs/components"
	"github.com/go-gl/mathgl/mgl32"
)

func TestEntityCreation(t *testing.T) {
	world := ecs.NewWorld()
	entity := world.CreateEntity()

	if entity == ecs.NullEntity {
		t.Fatal("CreateEntity returned NullEntity")
	}

	if !world.IsAlive(entity) {
		t.Fatal("Created entity is not alive")
	}
}

func TestEntityDestruction(t *testing.T) {
	world := ecs.NewWorld()
	entity := world.CreateEntity()

	world.DestroyEntity(entity)

	if world.IsAlive(entity) {
		t.Fatal("Destroyed entity is still alive")
	}
}

func TestEntityReuse(t *testing.T) {
	world := ecs.NewWorld()

	entity1 := world.CreateEntity()
	world.DestroyEntity(entity1)

	entity2 := world.CreateEntity()

	if entity1 != entity2 {
		t.Error("Entity IDs not reused properly")
	}
}

func TestComponentAddGet(t *testing.T) {
	world := ecs.NewWorld()
	entity := world.CreateEntity()

	transform := components.Transform{
		Position: mgl32.Vec3{1, 2, 3},
	}

	ecs.AddComponent(world, entity, transform)

	retrieved, exists := ecs.GetComponent[components.Transform](world, entity)
	if !exists {
		t.Fatal("Component not found after adding")
	}

	if retrieved.Position != transform.Position {
		t.Error("Retrieved component has wrong data")
	}
}

func TestComponentRemove(t *testing.T) {
	world := ecs.NewWorld()
	entity := world.CreateEntity()

	ecs.AddComponent(world, entity, components.NewTransform())

	if !ecs.HasComponent[components.Transform](world, entity) {
		t.Fatal("Component not found after adding")
	}

	removed := ecs.RemoveComponent[components.Transform](world, entity)
	if !removed {
		t.Error("RemoveComponent returned false")
	}

	if ecs.HasComponent[components.Transform](world, entity) {
		t.Error("Component still exists after removal")
	}
}

func TestQueryWith(t *testing.T) {
	world := ecs.NewWorld()

	// Entity with Transform only
	e1 := world.CreateEntity()
	ecs.AddComponent(world, e1, components.NewTransform())

	// Entity with Transform + Physics
	e2 := world.CreateEntity()
	ecs.AddComponent(world, e2, components.NewTransform())
	ecs.AddComponent(world, e2, components.NewPhysics(10.0))

	// Query Transform + Physics
	query := world.Query().
		With(ecs.ComponentTypeTransform).
		With(ecs.ComponentTypePhysics).
		Build()

	entities := query.Entities()
	if len(entities) != 1 {
		t.Fatalf("Expected 1 entity, got %d", len(entities))
	}

	if entities[0] != e2 {
		t.Error("Query returned wrong entity")
	}
}

func TestQueryWithout(t *testing.T) {
	world := ecs.NewWorld()

	e1 := world.CreateEntity()
	ecs.AddComponent(world, e1, components.NewTransform())

	e2 := world.CreateEntity()
	ecs.AddComponent(world, e2, components.NewTransform())
	ecs.AddComponent(world, e2, components.NewPhysics(10.0))

	// Query Transform WITHOUT Physics
	query := world.Query().
		With(ecs.ComponentTypeTransform).
		Without(ecs.ComponentTypePhysics).
		Build()

	entities := query.Entities()
	if len(entities) != 1 {
		t.Fatalf("Expected 1 entity, got %d", len(entities))
	}

	if entities[0] != e1 {
		t.Error("Query returned wrong entity")
	}
}

func TestWorldClear(t *testing.T) {
	world := ecs.NewWorld()

	// Create some entities
	for i := 0; i < 10; i++ {
		entity := world.CreateEntity()
		ecs.AddComponent(world, entity, components.NewTransform())
	}

	if world.EntityCount() != 10 {
		t.Fatalf("Expected 10 entities, got %d", world.EntityCount())
	}

	world.Clear()

	if world.EntityCount() != 0 {
		t.Error("World not empty after Clear()")
	}
}

func TestComponentStorage(t *testing.T) {
	storage := ecs.NewComponentStorage[components.Transform]()

	entity := ecs.Entity(1)
	transform := components.Transform{
		Position: mgl32.Vec3{5, 10, 15},
	}

	storage.Add(entity, transform)

	if !storage.Has(entity) {
		t.Fatal("Storage doesn't have entity after Add")
	}

	retrieved, exists := storage.Get(entity)
	if !exists {
		t.Fatal("Get returned false for existing component")
	}

	if retrieved.Position != transform.Position {
		t.Error("Retrieved component has wrong data")
	}

	storage.Remove(entity)

	if storage.Has(entity) {
		t.Error("Storage still has entity after Remove")
	}
}

func TestMultipleComponents(t *testing.T) {
	world := ecs.NewWorld()
	entity := world.CreateEntity()

	// Add multiple components
	ecs.AddComponent(world, entity, components.NewTransform())
	ecs.AddComponent(world, entity, components.NewPhysics(10.0))
	ecs.AddComponent(world, entity, components.NewModel("test.mdl"))

	if !ecs.HasComponent[components.Transform](world, entity) {
		t.Error("Missing Transform component")
	}
	if !ecs.HasComponent[components.Physics](world, entity) {
		t.Error("Missing Physics component")
	}
	if !ecs.HasComponent[components.Model](world, entity) {
		t.Error("Missing Model component")
	}

	// Remove one component
	ecs.RemoveComponent[components.Physics](world, entity)

	if !ecs.HasComponent[components.Transform](world, entity) {
		t.Error("Transform removed when it shouldn't be")
	}
	if ecs.HasComponent[components.Physics](world, entity) {
		t.Error("Physics not removed")
	}
	if !ecs.HasComponent[components.Model](world, entity) {
		t.Error("Model removed when it shouldn't be")
	}
}
