package entity

import (
	"testing"
)

func TestPropPhysicsOverride_Classname(t *testing.T) {
	sut := PropPhysicsOverride{}
	if sut.Classname() != "prop_physics_override" {
		t.Errorf("expected classname: prop_physics_override, but got: %s", sut.Classname())
	}
}
