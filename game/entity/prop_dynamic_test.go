package entity

import (
	"testing"
)

func TestPropDynamic_Classname(t *testing.T) {
	sut := PropDynamic{}
	if sut.Classname() != "prop_dynamic" {
		t.Errorf("expected classname: prop_dynamic, but got: %s", sut.Classname())
	}
}
