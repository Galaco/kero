package entity

import (
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/framework/graphics/mesh"
	"github.com/galaco/source-tools-common/entity"
	"github.com/go-gl/mathgl/mgl32"
)

// IEntity
type IEntity interface {
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
	// Transform contains the entity's representation in 3d space (non-renderable entities still have these properties)
	Transform() *graphics.Transform

	Model() *mesh.ModelInstance
	AttachModel(m *mesh.Model)
}

// Entity is a common base that most entities can be based upon
type Entity struct {
	entity.Entity
	// Transform contains the entity's representation in 3d space (non-renderable entities still have these properties)
	transform graphics.Transform
	// Class contains the entity's classname (e.g. func_movelinear)
	class string
	// Name contains the entity's targetname
	name string

	model *mesh.ModelInstance
}

// Classname returns the entity classname
func (e *Entity) Classname() string {
	return e.class
}

// Targetname returns the entity targername
func (e *Entity) Targetname() string {
	return e.name
}

// Transform contains the entity's representation in 3d space (non-renderable entities still have these properties)
func (e *Entity) Transform() *graphics.Transform {
	return &e.transform
}

// Origin returns the entity position in the world
func (e *Entity) Origin() mgl32.Vec3 {
	return e.transform.Position
}

// Angles returns the entity orientation in the world
func (e *Entity) Angles() mgl32.Vec3 {
	return e.transform.Rotation
}

// Properties returns all the entity's key-values as a linked list
func (e *Entity) Properties() *entity.EPair {
	return e.EPairs
}

// Think runs entity specific logic based on the elapsed time of the current frame
func (e *Entity) Think(dt float64) {

}

func (e *Entity) Model() *mesh.ModelInstance {
	return e.model
}
func (e *Entity) AttachModel(m *mesh.Model) {
	e.model = &mesh.ModelInstance{
		Model: m,
	}
}

// NewEntityBaseFromLib returns a new entity
func NewEntityBaseFromLib(e entity.Entity) *Entity {
	return &Entity{
		Entity: e,
		transform: graphics.Transform{
			Position: e.VectorForKey("origin"),
			Rotation: e.VectorForKey("angles"),
		},
		class: e.ValueForKey("classname"),
		name:  e.ValueForKey("targetname"),
	}
}

// NewEntityBase returns a new base entity
func NewEntityBase(classname, targetname string, transform graphics.Transform) *Entity {
	return &Entity{
		transform: transform,
		class:     classname,
		name:      targetname,
	}
}
