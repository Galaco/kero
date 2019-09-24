package entity

import (
	"reflect"
	"testing"
)

func TestPropRagdoll_Classname(t *testing.T) {
	sut := PropRagdoll{}
	if sut.Classname() != "prop_ragdoll" {
		t.Errorf("expected classname: prop_ragdoll, but got: %s", sut.Classname())
	}
}

func TestPropRagdoll_New(t *testing.T) {
	sut := &PropRagdoll{}

	actual := sut.New()
	if reflect.TypeOf(actual) != reflect.TypeOf(sut) {
		t.Errorf("unexpected type returned from New. Expected: %s, but received: %s", reflect.TypeOf(sut), reflect.TypeOf(actual))
	}
}
