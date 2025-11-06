package event

// IDispatcher
type IDispatcher interface {
	Initialize()
	DispatchLegacy(message Dispatchable)
	Dispatch(name Type, value interface{})
	AddListener(s Receiveable)
	CancelPending()
}

var masterDispatcher Dispatcher

// Deprecated: Use Engine.EventBus() instead of this global singleton.
// This function will be removed in a future version.
// For new code, pass EventBus as an explicit dependency via constructors.
func Get() *Dispatcher {
	if masterDispatcher.listeners == nil {
		masterDispatcher.Initialize()
	}
	return &masterDispatcher
}

// Dispatcher manages game events
// Now supports both legacy interface{}-based events and new type-safe generic events
type Dispatcher struct {
	messages    []Dispatchable
	newMessages []Dispatchable
	listeners   map[Type][]Receiveable

	// Type-safe event system (Phase 3)
	*DeferredDispatcher
}

// DispatchLegacy queues a message to be sent to listeners
func (eventBus *Dispatcher) DispatchLegacy(message Dispatchable) {
	if _, ok := eventBus.listeners[message.Type()]; ok {
		for _, cb := range eventBus.listeners[message.Type()] {
			cb(message)
		}
	}
}

// Dispatch sends a message to all listeners of the specified Type
func (eventBus *Dispatcher) Dispatch(name Type, message interface{}) {
	if _, ok := eventBus.listeners[name]; ok {
		for _, cb := range eventBus.listeners[name] {
			cb(message)
		}
	}
}

// CancelPending wipes the current queue.
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
	eventBus.DeferredDispatcher = NewDeferredDispatcher()
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		listeners:           map[Type][]Receiveable{},
		DeferredDispatcher:  NewDeferredDispatcher(),
	}
}

// ============================================================================
// Type-Safe Event Methods (Phase 3)
// ============================================================================
// Note: These are convenience wrappers around the free functions in typed_bus.go and phases.go

// RegisterTyped adds a type-safe event listener.
// This is a convenience wrapper for cleaner syntax.
//
// Example:
//   handle := bus.RegisterTyped(func(e messages.ChangeLevelEvent) {
//       fmt.Printf("Loading map: %s\n", e.MapName)
//   })
func RegisterTypedEvent[T any](bus *Dispatcher, handler EventListener[T]) EventHandle {
	return Register(bus.TypedDispatcher, handler)
}

// DispatchTyped sends a typed event to all registered listeners immediately.
//
// Example:
//   DispatchTyped(bus, messages.EngineQuitEvent{})
func DispatchTyped[T any](bus *Dispatcher, event T) {
	Dispatch(bus.TypedDispatcher, event)
}

// DispatchTypedDeferred queues a typed event to be processed during a specific phase.
//
// Example:
//   DispatchTypedDeferred(bus, PhasePostUpdate, messages.ChangeLevelEvent{MapName: "newmap"})
func DispatchTypedDeferred[T any](bus *Dispatcher, phase EventPhase, event T) {
	DispatchDeferred(bus.DeferredDispatcher, phase, event)
}

// UnregisterTyped removes a type-safe event listener using its handle.
func (eventBus *Dispatcher) UnregisterTyped(handle EventHandle) {
	eventBus.TypedDispatcher.Unregister(handle)
}
