package entity

import (
	"testing"
)

func TestPropDynamicOrnament_Classname(t *testing.T) {
	sut := PropDynamicOrnament{}
	if sut.Classname() != "prop_dynamic_ornament" {
		t.Errorf("expected classname: prop_dynamic_ornament, but got: %s", sut.Classname())
	}
}
