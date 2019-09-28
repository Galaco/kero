package event

var eventBus Dispatcher

// Singleton returns the global event Dispatcher.
func Singleton() *Dispatcher {
	return &eventBus
}

type Dispatcher struct {
	messages    []Dispatchable
	newMessages []Dispatchable
	systems     []receiveable
}

// ProcessMessages
func (b *Dispatcher) ProcessMessages() {
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
func (b *Dispatcher) Dispatch(message Dispatchable) {
	b.newMessages = append(b.newMessages, message)
}

func (b *Dispatcher) ClearQueue() {
	b.newMessages = make([]Dispatchable, 0)
}

// RegisterSystem
func (b *Dispatcher) RegisterSystem(s receiveable) {
	b.systems = append(b.systems, s)
}
