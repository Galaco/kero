package entity

import (
	"reflect"
	"testing"
)

func TestPropPhysicsMultiplayer_Classname(t *testing.T) {
	sut := PropPhysicsMultiplayer{}
	if sut.Classname() != "prop_physics_multiplayer" {
		t.Errorf("expected classname: prop_physics_multiplayer, but got: %s", sut.Classname())
	}
}

func TestPropPhysicsMultiplayer_New(t *testing.T) {
	sut := &PropPhysicsMultiplayer{}

	actual := sut.New()
	if reflect.TypeOf(actual) != reflect.TypeOf(sut) {
		t.Errorf("unexpected type returned from New. Expected: %s, but received: %s", reflect.TypeOf(sut), reflect.TypeOf(actual))
	}
}
