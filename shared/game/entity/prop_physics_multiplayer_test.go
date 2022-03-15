package entity

import (
	"testing"
)

func TestPropPhysicsMultiplayer_Classname(t *testing.T) {
	sut := PropPhysicsMultiplayer{}
	if sut.Classname() != "prop_physics_multiplayer" {
		t.Errorf("expected classname: prop_physics_multiplayer, but got: %s", sut.Classname())
	}
}
