package console

import "testing"

func TestAddOutputPipe(t *testing.T) {
	sut := false
	AddOutputPipe(func(f LogLevel, a interface{}) {
		sut = true
	})

	PrintString(LevelInfo, "foo")

	if sut != true {
		t.Error("added output pipe was not executed")
	}
}
