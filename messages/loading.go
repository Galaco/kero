package messages

import (
	"github.com/galaco/kero/event/message"
	"github.com/galaco/kero/framework/valve"
)

const (
	TypeLoadingLevelParsed = message.Type("LoadingLevelParsed")
)

type LoadingLevelParsed struct {
	level *valve.Bsp
}

func (msg *LoadingLevelParsed) Type() message.Type {
	return TypeLoadingLevelParsed
}

func (msg *LoadingLevelParsed) Level()*valve.Bsp {
	return msg.level
}

func NewLoadingLevelParsed(level *valve.Bsp) *LoadingLevelParsed {
	return &LoadingLevelParsed{
		level: level,
	}
}
