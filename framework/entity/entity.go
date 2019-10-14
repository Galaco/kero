package entity

import (
	graphics3d "github.com/galaco/kero/framework/graphics/3d"
	"github.com/galaco/source-tools-common/entity"
	"github.com/go-gl/mathgl/mgl32"
)

// Entity
type Entity interface {
	// Classname is the entity type
	Classname() string
	// Targetname is the name.  This is not unique
	Targetname() string
	// Origin is position in the world (x,y,z)
	Origin() mgl32.Vec3
	// Angles is the orientation in the world (x,y,z)
	Angles() mgl32.Vec3
	// Think updates this entity based on elapsed time since last update
	Think(dt float64)
	// ValueForKey provides the raw value of an entity key
	ValueForKey(key string) string
	// VectorForKey transforms a key-value into a 3d-vector
	VectorForKey(key string) mgl32.Vec3
	// IntForKey transforms a key-value into an int
	IntForKey(key string) int
	// FloatForKey transforms a key-value into a float
	FloatForKey(key string) float32
	// FloatForKeyWithDefault transforms a key-value into a float if possible, otherwise
	// uses a provided default
	FloatForKeyWithDefault(key string, defaultValue float32) float32
	// Properties returns a linked list of all entity key-values
	Properties() *entity.EPair
}

// EntityBase is a common base that most entities can be based upon
type EntityBase struct {
	entity.Entity
	Transform graphics3d.Transform
	Class     string
	Name      string
}

// Classname returns the entity classname
func (e *EntityBase) Classname() string {
	return e.Class
}

// Targetname returns the entity targername
func (e *EntityBase) Targetname() string {
	return e.Name
}

// Origin returns the entity position in the world
func (e *EntityBase) Origin() mgl32.Vec3 {
	return e.Transform.Position
}

// Angles returns the entity orientation in the world
func (e *EntityBase) Angles() mgl32.Vec3 {
	return e.Transform.Rotation
}

// Properties returns all the entity's key-values as a linked list
func (e *EntityBase) Properties() *entity.EPair {
	return e.EPairs
}

// Think runs entity specific logic based on the elapsed time of the current frame
func (e *EntityBase) Think(dt float64) {

}

// NewEntityBaseFromLib returns a new entity
func NewEntityBaseFromLib(e entity.Entity) *EntityBase {
	return &EntityBase{
		Entity: e,
		Transform: graphics3d.Transform{
			Position: e.VectorForKey("origin"),
			Rotation: e.VectorForKey("angles"),
		},
		Class: e.ValueForKey("classname"),
		Name:  e.ValueForKey("targetname"),
	}
}

// NewEntityBase returns a new base entity
func NewEntityBase(classname, targetname string, transform graphics3d.Transform) *EntityBase {
	return &EntityBase{
		Transform: transform,
		Class:     classname,
		Name:      targetname,
	}
}
