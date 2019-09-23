package event

import (
	"github.com/galaco/kero/event/message"
	"github.com/galaco/kero/systems"
)

var eventBus bus

// Singleton returns the global event bus.
func Singleton() *bus {
	return &eventBus
}

type bus struct {
	messages    []message.Dispatchable
	newMessages []message.Dispatchable
	systems     []systems.ISystem
}

// ProcessMessages
func (b *bus) ProcessMessages() {
	for _, m := range b.messages {
		for _, s := range b.systems {
			s.ProcessMessage(m)
			if len(b.messages) < 2 {
				b.messages = make([]message.Dispatchable, 0)
			} else {
				b.messages = b.messages[1:]
			}
		}
	}

	b.messages = b.newMessages
	b.newMessages = make([]message.Dispatchable, 0)
}

// Dispatch
func (b *bus) Dispatch(message message.Dispatchable) {
	b.newMessages = append(b.newMessages, message)
}

func (b *bus) ClearQueue() {
	b.newMessages = make([]message.Dispatchable, 0)
}

// RegisterSystem
func (b *bus) RegisterSystem(s systems.ISystem) {
	b.systems = append(b.systems, s)
}
