package ecs

import (
	"sync"
)

// World manages all entities and their components.
// It owns component storage and provides the API for entity/component operations.
type World struct {
	// Entity management
	entities     []EntityEntry
	freeEntities []Entity
	nextEntityID uint64
	entityMu     sync.RWMutex

	// Component storages (type-erased map)
	storages   map[ComponentType]interface{}
	storageMu  sync.RWMutex

	// Query cache for performance
	queryCache   map[string]*Query
	queryCacheMu sync.RWMutex
}

// NewWorld creates a new ECS world
func NewWorld() *World {
	world := &World{
		entities:   make([]EntityEntry, 1, 1024), // Reserve index 0 for NullEntity
		storages:   make(map[ComponentType]interface{}),
		queryCache: make(map[string]*Query),
	}
	// Initialize index 0 as the null entity (never alive)
	world.entities[0] = EntityEntry{
		ID:         NullEntity,
		Generation: 0,
		Alive:      false,
	}
	return world
}

// CreateEntity allocates a new entity ID
func (w *World) CreateEntity() Entity {
	w.entityMu.Lock()
	defer w.entityMu.Unlock()

	// Reuse freed entity IDs
	if len(w.freeEntities) > 0 {
		entity := w.freeEntities[len(w.freeEntities)-1]
		w.freeEntities = w.freeEntities[:len(w.freeEntities)-1]

		entry := &w.entities[entity]
		entry.Alive = true
		entry.Generation++
		return entity
	}

	// Allocate new entity (use slice index as entity ID)
	entityID := Entity(len(w.entities))
	w.entities = append(w.entities, EntityEntry{
		ID:         entityID,
		Generation: 0,
		Alive:      true,
	})

	return entityID
}

// DestroyEntity removes an entity and all its components
func (w *World) DestroyEntity(entity Entity) {
	w.entityMu.Lock()
	defer w.entityMu.Unlock()

	if entity >= Entity(len(w.entities)) || !w.entities[entity].Alive {
		return
	}

	w.entities[entity].Alive = false
	w.freeEntities = append(w.freeEntities, entity)

	// Remove all components from all storages
	w.storageMu.RLock()
	defer w.storageMu.RUnlock()

	for _, storage := range w.storages {
		// Type-erased removal using interface
		if remover, ok := storage.(interface{ Remove(Entity) bool }); ok {
			remover.Remove(entity)
		}
	}
}

// IsAlive checks if an entity is currently valid
func (w *World) IsAlive(entity Entity) bool {
	w.entityMu.RLock()
	defer w.entityMu.RUnlock()

	if entity >= Entity(len(w.entities)) {
		return false
	}
	return w.entities[entity].Alive
}

// EntityCount returns the number of alive entities
func (w *World) EntityCount() int {
	w.entityMu.RLock()
	defer w.entityMu.RUnlock()

	count := 0
	for i := range w.entities {
		if w.entities[i].Alive {
			count++
		}
	}
	return count
}

// AddComponent adds a component to an entity.
// This is a free function to support generic type parameters.
func AddComponent[T Component](w *World, entity Entity, component T) {
	componentType := GetComponentType[T]()
	if componentType == ComponentTypeInvalid {
		panic("component type not registered")
	}

	storage := getOrCreateStorage[T](w, componentType)
	storage.Add(entity, component)
}

// GetComponent retrieves a component from an entity.
// Returns (component, true) if found, (nil, false) otherwise.
func GetComponent[T Component](w *World, entity Entity) (*T, bool) {
	componentType := GetComponentType[T]()
	if componentType == ComponentTypeInvalid {
		return nil, false
	}

	storage := getStorage[T](w, componentType)
	if storage == nil {
		return nil, false
	}
	return storage.Get(entity)
}

// RemoveComponent removes a component from an entity
func RemoveComponent[T Component](w *World, entity Entity) bool {
	componentType := GetComponentType[T]()
	if componentType == ComponentTypeInvalid {
		return false
	}

	storage := getStorage[T](w, componentType)
	if storage == nil {
		return false
	}
	return storage.Remove(entity)
}

// HasComponent checks if an entity has a specific component
func HasComponent[T Component](w *World, entity Entity) bool {
	_, has := GetComponent[T](w, entity)
	return has
}

// GetStorage retrieves the storage for a component type (type-safe).
// This is a free function to support generic type parameters.
func GetStorage[T Component](w *World, componentType ComponentType) *ComponentStorage[T] {
	return getStorage[T](w, componentType)
}

// getOrCreateStorage retrieves or creates storage for a component type
func getOrCreateStorage[T Component](w *World, componentType ComponentType) *ComponentStorage[T] {
	w.storageMu.Lock()
	defer w.storageMu.Unlock()

	if storage, exists := w.storages[componentType]; exists {
		return storage.(*ComponentStorage[T])
	}

	storage := NewComponentStorage[T]()
	w.storages[componentType] = storage
	return storage
}

// getStorage retrieves storage for a component type (returns nil if doesn't exist)
func getStorage[T Component](w *World, componentType ComponentType) *ComponentStorage[T] {
	w.storageMu.RLock()
	defer w.storageMu.RUnlock()

	if storage, exists := w.storages[componentType]; exists {
		return storage.(*ComponentStorage[T])
	}
	return nil
}

// getStorageByType retrieves type-erased storage by component type
func (w *World) getStorageByType(componentType ComponentType) interface{} {
	w.storageMu.RLock()
	defer w.storageMu.RUnlock()

	return w.storages[componentType]
}

// Clear removes all entities and components
func (w *World) Clear() {
	w.entityMu.Lock()
	// Reset to just the null entity at index 0
	w.entities = w.entities[:1]
	w.entities[0] = EntityEntry{
		ID:         NullEntity,
		Generation: 0,
		Alive:      false,
	}
	w.freeEntities = w.freeEntities[:0]
	w.nextEntityID = 0
	w.entityMu.Unlock()

	w.storageMu.Lock()
	for _, storage := range w.storages {
		if clearer, ok := storage.(interface{ Clear() }); ok {
			clearer.Clear()
		}
	}
	w.storageMu.Unlock()

	w.queryCacheMu.Lock()
	w.queryCache = make(map[string]*Query)
	w.queryCacheMu.Unlock()
}
