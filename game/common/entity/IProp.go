package entity

import "github.com/galaco/lambda-core/model"

// IProp Base renderable prop interface
type IProp interface {
	Model() *model.Model
	SetModel(model *model.Model)
}

// PropBase is a minimal renderable prop entity
type PropBase struct {
	model *model.Model
}

// SetModel
func (prop *PropBase) SetModel(model *model.Model) {
	prop.model = model
}

// Model
func (prop *PropBase) Model() *model.Model {
	return prop.model
}
