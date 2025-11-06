package ecs

import "sync"

// ComponentStorage stores components of a specific type.
// Uses sparse set data structure for fast iteration and lookup.
//
// Components are stored in a dense array for cache-friendly iteration.
// A sparse map provides O(1) lookup from Entity → component index.
type ComponentStorage[T Component] struct {
	components []T             // Dense array of components
	entities   []Entity        // Parallel array: which entity owns each component
	sparse     map[Entity]int  // Entity → index in dense arrays
	mu         sync.RWMutex
}

// NewComponentStorage creates a new storage for a component type
func NewComponentStorage[T Component]() *ComponentStorage[T] {
	return &ComponentStorage[T]{
		components: make([]T, 0, 256),
		entities:   make([]Entity, 0, 256),
		sparse:     make(map[Entity]int),
	}
}

// Add adds a component to an entity
func (s *ComponentStorage[T]) Add(entity Entity, component T) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already exists
	if _, exists := s.sparse[entity]; exists {
		panic("component already exists on entity")
	}

	index := len(s.components)
	s.components = append(s.components, component)
	s.entities = append(s.entities, entity)
	s.sparse[entity] = index
}

// Get retrieves a component for an entity
func (s *ComponentStorage[T]) Get(entity Entity) (*T, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	index, exists := s.sparse[entity]
	if !exists {
		return nil, false
	}
	return &s.components[index], true
}

// Has checks if an entity has this component
func (s *ComponentStorage[T]) Has(entity Entity) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.sparse[entity]
	return exists
}

// Remove removes a component from an entity.
// Uses swap-and-pop to maintain dense packing.
func (s *ComponentStorage[T]) Remove(entity Entity) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	index, exists := s.sparse[entity]
	if !exists {
		return false
	}

	// Swap with last element (maintains dense packing)
	lastIndex := len(s.components) - 1
	if index != lastIndex {
		s.components[index] = s.components[lastIndex]
		s.entities[index] = s.entities[lastIndex]
		s.sparse[s.entities[index]] = index
	}

	// Remove last element
	s.components = s.components[:lastIndex]
	s.entities = s.entities[:lastIndex]
	delete(s.sparse, entity)

	return true
}

// Len returns the number of components stored
func (s *ComponentStorage[T]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.components)
}

// GetEntities returns all entities that have this component
func (s *ComponentStorage[T]) GetEntities() []Entity {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to avoid concurrent modification issues
	result := make([]Entity, len(s.entities))
	copy(result, s.entities)
	return result
}

// GetComponentsSlice returns the internal components slice (read-only).
// UNSAFE: Caller must hold read lock and not modify the slice.
func (s *ComponentStorage[T]) GetComponentsSlice() []T {
	return s.components
}

// GetEntitiesSlice returns the internal entities slice (read-only).
// UNSAFE: Caller must hold read lock and not modify the slice.
func (s *ComponentStorage[T]) GetEntitiesSlice() []Entity {
	return s.entities
}

// RLock locks the storage for reading
func (s *ComponentStorage[T]) RLock() {
	s.mu.RLock()
}

// RUnlock unlocks the storage after reading
func (s *ComponentStorage[T]) RUnlock() {
	s.mu.RUnlock()
}

// Clear removes all components
func (s *ComponentStorage[T]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.components = s.components[:0]
	s.entities = s.entities[:0]
	s.sparse = make(map[Entity]int)
}
