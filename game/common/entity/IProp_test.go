package entity

import (
	"github.com/galaco/lambda-core/model"
	"testing"
)

func TestPropBase_Model(t *testing.T) {
	sut := PropBase{}
	if sut.Model() != nil {
		t.Error("model was set, but should not be")
	}

	mod := &model.Model{}
	sut.SetModel(mod)

	if sut.Model() != mod {
		t.Errorf("set mode l does not match expected")
	}
}

func TestPropBase_SetModel(t *testing.T) {
	sut := PropBase{}
	if sut.Model() != nil {
		t.Error("model was set, but should not be")
	}

	mod := &model.Model{}
	sut.SetModel(mod)

	if sut.Model() != mod {
		t.Errorf("set mode l does not match expected")
	}
}
