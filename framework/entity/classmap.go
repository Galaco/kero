package entity

import "sync"

var classMap entityClassMapper

// entityClassMapper provides a reflection-like construct for creating
// new entity objects of a known Classname.
// The idea behind this was to remove the need to slow, difficult
// reflection. Instead, it is up to defined entity types to provide a means
// to create a new instance of its own type; this class being used to provide
// a gateway to manage that mapping.
// Note: this class is somewhat memory costly, as a single unmodified instance for every
// mapped type is required for storage. Templated functions would probably solve this
// problem better if they existed, and the plan was to avoid actual reflection
// where possible.
type entityClassMapper struct {
	entityMap map[string]IEntity
	mut       sync.Mutex
}

// find creates a new IEntity of the specified
// Classname.
func (classMap *entityClassMapper) find(classname string) IEntity {
	classMap.mut.Lock()
	defer classMap.mut.Unlock()
	if classMap.entityMap[classname] != nil {
		return classMap.entityMap[classname]
	}
	return nil
}

// Deprecated: Use Engine.EntityRegistry().RegisterClass() instead.
// This function will be removed in a future version.
// For new code, pass Registry as an explicit dependency via constructors.
func RegisterClass(entity IEntity) {
	if classMap.entityMap == nil {
		classMap.entityMap = map[string]IEntity{}
	}

	classMap.mut.Lock()
	classMap.entityMap[entity.Classname()] = entity
	classMap.mut.Unlock()
}

// Deprecated: Use Engine.EntityRegistry().New() instead.
// This function will be removed in a future version.
func New(classname string) IEntity {
	return classMap.find(classname)
}
