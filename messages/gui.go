package messages

import (
	"github.com/galaco/kero/framework/event"
)

const (
	// TypeChangeLevel
	TypeChangeLevel = event.Type("ChangeLevel")
)

// ChangeLevel
type ChangeLevel struct {
	levelName string
}

// Type
func (msg *ChangeLevel) Type() event.Type {
	return TypeChangeLevel
}

// LevelName
func (msg *ChangeLevel) LevelName() string {
	return msg.levelName
}

// NewChangeLevel
func NewChangeLevel(levelName string) *ChangeLevel {
	return &ChangeLevel{
		levelName: levelName,
	}
}
