package entity

import (
	"testing"
)

func TestPropDynamicOverride_Classname(t *testing.T) {
	sut := PropDynamicOverride{}
	if sut.Classname() != "prop_dynamic_override" {
		t.Errorf("expected classname: prop_dynamic_override, but got: %s", sut.Classname())
	}
}
