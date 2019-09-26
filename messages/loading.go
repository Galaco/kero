package messages

import (
	"github.com/galaco/kero/event"
	"github.com/galaco/kero/valve"
)

const (
	TypeLoadingLevelParsed = event.Type("LoadingLevelParsed")
)

type LoadingLevelParsed struct {
	level *valve.Bsp
}

func (msg *LoadingLevelParsed) Type() event.Type {
	return TypeLoadingLevelParsed
}

func (msg *LoadingLevelParsed) Level() *valve.Bsp {
	return msg.level
}

func NewLoadingLevelParsed(level *valve.Bsp) *LoadingLevelParsed {
	return &LoadingLevelParsed{
		level: level,
	}
}
