package event

var eventBus Dispatcher

type Dispatcher struct {
	messages    []Dispatchable
	newMessages []Dispatchable
	systems     []receiveable
}

// ProcessMessages
func ProcessMessages() {
	for _, m := range eventBus.messages {
		for _, s := range eventBus.systems {
			s.ProcessMessage(m)
			if len(eventBus.messages) < 2 {
				eventBus.messages = make([]Dispatchable, 0)
			} else {
				eventBus.messages = eventBus.messages[1:]
			}
		}
	}

	eventBus.messages = eventBus.newMessages
	eventBus.newMessages = make([]Dispatchable, 0)
}

// Dispatch
func Dispatch(message Dispatchable) {
	eventBus.newMessages = append(eventBus.newMessages, message)
}

func ClearQueue() {
	eventBus.newMessages = make([]Dispatchable, 0)
}

// RegisterSystem
func RegisterSystem(s receiveable) {
	eventBus.systems = append(eventBus.systems, s)
}
