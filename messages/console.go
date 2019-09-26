package messages

import (
	"github.com/galaco/kero/event"
	"github.com/galaco/kero/framework/console"
)

const (
	TypeConsoleMessage = event.Type("ConsoleMessage")
)

type ConsoleMessage struct {
	level   console.LogLevel
	message string
}

func (msg *ConsoleMessage) Type() event.Type {
	return TypeConsoleMessage
}

func (msg *ConsoleMessage) Level() console.LogLevel {
	return msg.level
}

func (msg *ConsoleMessage) Message() string {
	return msg.message
}

func NewConsoleMessage(level console.LogLevel, message string) *ConsoleMessage {
	return &ConsoleMessage{
		level:   level,
		message: message,
	}
}
