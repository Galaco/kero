package entity

import (
	"reflect"
	"testing"
)

func TestPropDynamic_Classname(t *testing.T) {
	sut := PropDynamic{}
	if sut.Classname() != "prop_dynamic" {
		t.Errorf("expected classname: prop_dynamic, but got: %s", sut.Classname())
	}
}

func TestPropDynamic_New(t *testing.T) {
	sut := &PropDynamic{}

	actual := sut.New()
	if reflect.TypeOf(actual) != reflect.TypeOf(sut) {
		t.Errorf("unexpected type returned from New. Expected: %s, but received: %s", reflect.TypeOf(sut), reflect.TypeOf(actual))
	}
}
