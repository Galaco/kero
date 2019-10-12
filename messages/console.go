package messages

import (
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/console"
)

const (
	// TypeConsoleMessage
	TypeConsoleMessage = event.Type("ConsoleMessage")
)

// ConsoleMessage
type ConsoleMessage struct {
	level   console.LogLevel
	message string
}

// Type
func (msg *ConsoleMessage) Type() event.Type {
	return TypeConsoleMessage
}

// Level
func (msg *ConsoleMessage) Level() console.LogLevel {
	return msg.level
}

// Message
func (msg *ConsoleMessage) Message() string {
	return msg.message
}

// NewConsoleMessage
func NewConsoleMessage(level console.LogLevel, message string) *ConsoleMessage {
	return &ConsoleMessage{
		level:   level,
		message: message,
	}
}
