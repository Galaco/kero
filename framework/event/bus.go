package event

type IDispatcher interface {
	Initialize()
	DispatchLegacy(message Dispatchable)
	Dispatch(name Type, value interface{})
	AddListener(s Receiveable)
	CancelPending()
}

var masterDispatcher Dispatcher

func Get() *Dispatcher {
	if masterDispatcher.listeners == nil {
		masterDispatcher.Initialize()
	}
	return &masterDispatcher
}

// Dispatcher manages game events
type Dispatcher struct {
	messages    []Dispatchable
	newMessages []Dispatchable
	listeners   map[Type][]Receiveable
}

// ProcessMessages loops through all stored messages and dispatches
// them to listeners
//func (eventBus *Dispatcher) ProcessMessages() {
//	for _, m := range eventBus.messages {
//		for _, s := range eventBus.listeners {
//			s.ProcessMessage(m)
//			if len(eventBus.messages) < 2 {
//				eventBus.messages = make([]Dispatchable, 0)
//			} else {
//				eventBus.messages = eventBus.messages[1:]
//			}
//		}
//	}
//
//	eventBus.messages = eventBus.newMessages
//	eventBus.newMessages = make([]Dispatchable, 0)
//}

// DispatchLegacy queues a message to be sent to listeners
func (eventBus *Dispatcher) DispatchLegacy(message Dispatchable) {
	if _, ok := eventBus.listeners[message.Type()]; ok {
		for _, cb := range eventBus.listeners[message.Type()] {
			cb(message)
		}
	}
}

func (eventBus *Dispatcher) Dispatch(name Type, message interface{}) {
	if _, ok := eventBus.listeners[name]; ok {
		for _, cb := range eventBus.listeners[name] {
			cb(message)
		}
	}
}

// ClearQueue wipes the current queue.
// This should be used with care.
func (eventBus *Dispatcher) CancelPending() {
	eventBus.newMessages = make([]Dispatchable, 0)
}

// AddListener adds a listener for events.
func (eventBus *Dispatcher) AddListener(message Type, s Receiveable) {
	if _, ok := eventBus.listeners[message]; ok {
		eventBus.listeners[message] = append(eventBus.listeners[message], s)
	} else {
		eventBus.listeners[message] = []Receiveable{s}
	}
}

func (eventBus *Dispatcher) Initialize() {
	eventBus.listeners = map[Type][]Receiveable{}
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		listeners: map[Type][]Receiveable{},
	}
}
