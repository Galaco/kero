package ecs

// Entity is a unique identifier for an entity in the world.
// It's just an ID - all data is stored in components.
type Entity uint64

const (
	// NullEntity represents an invalid/non-existent entity
	NullEntity Entity = 0
)

// EntityEntry tracks the lifecycle of an entity
type EntityEntry struct {
	ID         Entity
	Generation uint32 // Incremented each time the ID is reused
	Alive      bool
}

// IsAlive returns true if this entity ID is currently valid
func (e EntityEntry) IsAlive() bool {
	return e.Alive
}

// IsNull returns true if this is the null entity
func (e Entity) IsNull() bool {
	return e == NullEntity
}
