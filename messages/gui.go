package messages

import (
	"github.com/galaco/kero/event"
)

const (
	TypeChangeLevel = event.Type("ChangeLevel")
)

type ChangeLevel struct {
	levelName string
}

func (msg *ChangeLevel) Type() event.Type {
	return TypeChangeLevel
}

func (msg *ChangeLevel) LevelName() string {
	return msg.levelName
}

func NewChangeLevel(levelName string) *ChangeLevel {
	return &ChangeLevel{
		levelName: levelName,
	}
}
