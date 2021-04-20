package entity

import (
	"testing"
)

func TestInfoPlayerStart_Classname(t *testing.T) {
	sut := InfoPlayerStart{}
	if sut.Classname() != "info_player_start" {
		t.Errorf("expected classname: info_player_start, but got: %s", sut.Classname())
	}
}
