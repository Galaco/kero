package entity

import (
	"testing"
)

func TestPropDoorRotating_Classname(t *testing.T) {
	sut := PropDoorRotating{}
	if sut.Classname() != "prop_door_rotating" {
		t.Errorf("expected classname: prop_door_rotating, but got: %s", sut.Classname())
	}
}
