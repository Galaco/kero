package event

var eventBus bus

// Singleton returns the global event bus.
func Singleton() *bus {
	return &eventBus
}

type bus struct {
	messages    []Dispatchable
	newMessages []Dispatchable
	systems     []receiveable
}

// ProcessMessages
func (b *bus) ProcessMessages() {
	for _, m := range b.messages {
		for _, s := range b.systems {
			s.ProcessMessage(m)
			if len(b.messages) < 2 {
				b.messages = make([]Dispatchable, 0)
			} else {
				b.messages = b.messages[1:]
			}
		}
	}

	b.messages = b.newMessages
	b.newMessages = make([]Dispatchable, 0)
}

// Dispatch
func (b *bus) Dispatch(message Dispatchable) {
	b.newMessages = append(b.newMessages, message)
}

func (b *bus) ClearQueue() {
	b.newMessages = make([]Dispatchable, 0)
}

// RegisterSystem
func (b *bus) RegisterSystem(s receiveable) {
	b.systems = append(b.systems, s)
}
