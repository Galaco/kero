package entity

import "sync"

// Registry provides a registry for entity classes.
// This replaces the global classMap singleton with an explicit instance.
// Entity classes are registered by classname, allowing creation of new
// entity instances by classname lookup at runtime.
type Registry struct {
	entityMap map[string]IEntity
	mu        sync.RWMutex
}

// NewRegistry creates a new entity class registry
func NewRegistry() *Registry {
	return &Registry{
		entityMap: make(map[string]IEntity),
	}
}

// RegisterClass adds an entity type to the registry.
// The entity's Classname() is used as the key.
// After registration, new instances can be created via New(classname).
func (r *Registry) RegisterClass(entity IEntity) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entityMap[entity.Classname()] = entity
}

// New creates a new IEntity instance of the specified classname.
// Returns nil if the classname is not registered.
func (r *Registry) New(classname string) IEntity {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if entity, ok := r.entityMap[classname]; ok {
		return entity
	}
	return nil
}

// Has checks if a classname is registered
func (r *Registry) Has(classname string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.entityMap[classname]
	return ok
}

// Count returns the number of registered entity classes
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.entityMap)
}
