package entity

import (
	"testing"
)

func TestPropPhysics_Classname(t *testing.T) {
	sut := PropPhysics{}
	if sut.Classname() != "prop_physics" {
		t.Errorf("expected classname: prop_physics, but got: %s", sut.Classname())
	}
}
