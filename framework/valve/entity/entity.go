package entity

import (
	graphics3d "github.com/galaco/kero/framework/graphics/3d"
	"github.com/galaco/source-tools-common/entity"
	"github.com/go-gl/mathgl/mgl32"
)

type Entity interface {
	Classname() string
	Targetname() string
	Origin() mgl32.Vec3
	Angles() mgl32.Vec3
	Think(dt float64)
	ValueForKey(key string) string
	VectorForKey(key string) mgl32.Vec3
	IntForKey(key string) int
	FloatForKey(key string) float32
	FloatForKeyWithDefault(key string, defaultValue float32) float32
	Properties() *entity.EPair
}

type EntityBase struct {
	entity.Entity
	Transform graphics3d.Transform
	Class     string
	Name      string
}

func (e *EntityBase) Classname() string {
	return e.Class
}

func (e *EntityBase) Targetname() string {
	return e.Name
}

func (e *EntityBase) Origin() mgl32.Vec3 {
	return e.Transform.Position
}

func (e *EntityBase) Angles() mgl32.Vec3 {
	return e.Transform.Rotation
}

func (e *EntityBase) Properties() *entity.EPair {
	return e.EPairs
}

func (e *EntityBase) Think(dt float64) {

}

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

func NewEntityBase(classname, targetname string, transform graphics3d.Transform) *EntityBase {
	return &EntityBase{
		Transform: transform,
		Class:     classname,
		Name:      targetname,
	}
}
