package event

var eventBus Dispatcher

// Dispatcher manages game events
type Dispatcher struct {
	messages    []Dispatchable
	newMessages []Dispatchable
	systems     []receiveable
}

// ProcessMessages loops through all stored messages and dispatches
// them to listeners
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

// Dispatch queues a message to be sent to listeners
func Dispatch(message Dispatchable) {
	eventBus.newMessages = append(eventBus.newMessages, message)
}

// ClearQueue wipes the current queue.
// This should be used with care.
func ClearQueue() {
	eventBus.newMessages = make([]Dispatchable, 0)
}

// AddListener adds a listener for events.
func AddListener(s receiveable) {
	eventBus.systems = append(eventBus.systems, s)
}
