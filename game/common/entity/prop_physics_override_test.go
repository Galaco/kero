package entity

import (
	"reflect"
	"testing"
)

func TestPropPhysicsOverride_Classname(t *testing.T) {
	sut := PropPhysicsOverride{}
	if sut.Classname() != "prop_physics_override" {
		t.Errorf("expected classname: prop_physics_override, but got: %s", sut.Classname())
	}
}

func TestPropPhysicsOverride_New(t *testing.T) {
	sut := &PropPhysicsOverride{}

	actual := sut.New()
	if reflect.TypeOf(actual) != reflect.TypeOf(sut) {
		t.Errorf("unexpected type returned from New. Expected: %s, but received: %s", reflect.TypeOf(sut), reflect.TypeOf(actual))
	}
}
