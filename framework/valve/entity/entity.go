package entity

import (
	graphics3d "github.com/galaco/kero/framework/graphics/3d"
	"github.com/go-gl/mathgl/mgl32"
)

type Entity interface {
	Classname() string
	Targetname() string
	Origin() mgl32.Vec3
	Angles() mgl32.Vec3
	Think(dt float64)
}

type EntityBase struct {
	Transform graphics3d.Transform
	Class string
	Name string
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

func (e *EntityBase) Think(dt float64) {

}

func NewEntityBase(classname, targetname string, transform graphics3d.Transform) EntityBase {
	return EntityBase{
		Transform: transform,
		Class:     classname,
		Name:      targetname,
	}
}