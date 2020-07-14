package mesh

import (
	"reflect"
	"testing"
)

func TestNewCube(t *testing.T) {
	sut := NewCube()
	if reflect.TypeOf(sut) != reflect.TypeOf(&Cube{}) {
		t.Error("unexpected value returned when creating Cube")
	}
}
