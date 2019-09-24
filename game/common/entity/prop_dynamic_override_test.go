package entity

import (
	"reflect"
	"testing"
)

func TestPropDynamicOverride_Classname(t *testing.T) {
	sut := PropDynamicOverride{}
	if sut.Classname() != "prop_dynamic_override" {
		t.Errorf("expected classname: prop_dynamic_override, but got: %s", sut.Classname())
	}
}

func TestPropDynamicOverride_New(t *testing.T) {
	sut := &PropDynamicOverride{}

	actual := sut.New()
	if reflect.TypeOf(actual) != reflect.TypeOf(sut) {
		t.Errorf("unexpected type returned from New. Expected: %s, but received: %s", reflect.TypeOf(sut), reflect.TypeOf(actual))
	}
}
