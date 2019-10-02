package entity

import (
	"testing"
)

func TestPropRagdoll_Classname(t *testing.T) {
	sut := PropRagdoll{}
	if sut.Classname() != "prop_ragdoll" {
		t.Errorf("expected classname: prop_ragdoll, but got: %s", sut.Classname())
	}
}