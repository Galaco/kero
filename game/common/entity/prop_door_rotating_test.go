package entity

import (
	"reflect"
	"testing"
)

func TestPropDoorRotating_Classname(t *testing.T) {
	sut := PropDoorRotating{}
	if sut.Classname() != "prop_door_rotating" {
		t.Errorf("expected classname: prop_door_rotating, but got: %s", sut.Classname())
	}
}

func TestPropDoorRotating_New(t *testing.T) {
	sut := &PropDoorRotating{}

	actual := sut.New()
	if reflect.TypeOf(actual) != reflect.TypeOf(sut) {
		t.Errorf("unexpected type returned from New. Expected: %s, but received: %s", reflect.TypeOf(sut), reflect.TypeOf(actual))
	}
}
