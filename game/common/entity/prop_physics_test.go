package entity

import (
	"reflect"
	"testing"
)

func TestPropPhysics_Classname(t *testing.T) {
	sut := PropPhysics{}
	if sut.Classname() != "prop_physics" {
		t.Errorf("expected classname: prop_physics, but got: %s", sut.Classname())
	}
}

func TestPropPhysics_New(t *testing.T) {
	sut := &PropPhysics{}

	actual := sut.New()
	if reflect.TypeOf(actual) != reflect.TypeOf(sut) {
		t.Errorf("unexpected type returned from New. Expected: %s, but received: %s", reflect.TypeOf(sut), reflect.TypeOf(actual))
	}
}
