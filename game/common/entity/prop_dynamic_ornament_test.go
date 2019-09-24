package entity

import (
	"reflect"
	"testing"
)

func TestPropDynamicOrnament_Classname(t *testing.T) {
	sut := PropDynamicOrnament{}
	if sut.Classname() != "prop_dynamic_ornament" {
		t.Errorf("expected classname: prop_dynamic_ornament, but got: %s", sut.Classname())
	}
}

func TestPropDynamicOrnament_New(t *testing.T) {
	sut := &PropDynamicOrnament{}

	actual := sut.New()
	if reflect.TypeOf(actual) != reflect.TypeOf(sut) {
		t.Errorf("unexpected type returned from New. Expected: %s, but received: %s", reflect.TypeOf(sut), reflect.TypeOf(actual))
	}
}
