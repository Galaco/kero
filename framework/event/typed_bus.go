package event

import (
	"reflect"
	"sync"
)

// EventListener is a generic type-safe event handler function.
// Instead of func(interface{}), handlers receive the actual event type.
type EventListener[T any] func(T)

// EventHandle allows unregistering event listeners.
type EventHandle struct {
	eventType reflect.Type
	id        uint64
}

// listenerList stores all handlers for a specific event type
type listenerList struct {
	handlers map[uint64]interface{} // map[id]EventListener[T]
	mu       sync.RWMutex
}

// TypedDispatcher provides type-safe event handling using Go generics.
// This eliminates runtime type assertion panics and enables IDE autocomplete.
type TypedDispatcher struct {
	listeners sync.Map // map[reflect.Type]*listenerList
	nextID    uint64
	mu        sync.Mutex
}

// NewTypedDispatcher creates a new type-safe event dispatcher
func NewTypedDispatcher() *TypedDispatcher {
	return &TypedDispatcher{
		nextID: 1,
	}
}

// Register adds a type-safe event listener.
// The type parameter T determines which events this listener receives.
// Returns a handle that can be used to unregister the listener.
//
// Example:
//   handle := Register(bus, func(e ChangeLevelEvent) {
//       fmt.Printf("Loading map: %s\n", e.MapName)
//   })
func Register[T any](bus *TypedDispatcher, handler EventListener[T]) EventHandle {
	eventType := reflect.TypeOf((*T)(nil)).Elem()

	bus.mu.Lock()
	id := bus.nextID
	bus.nextID++
	bus.mu.Unlock()

	// Get or create listener list for this type
	listInterface, _ := bus.listeners.LoadOrStore(eventType, &listenerList{
		handlers: make(map[uint64]interface{}),
	})
	list := listInterface.(*listenerList)

	list.mu.Lock()
	list.handlers[id] = handler
	list.mu.Unlock()

	return EventHandle{
		eventType: eventType,
		id:        id,
	}
}

// Unregister removes an event listener using its handle.
func (bus *TypedDispatcher) Unregister(handle EventHandle) {
	if listInterface, ok := bus.listeners.Load(handle.eventType); ok {
		list := listInterface.(*listenerList)
		list.mu.Lock()
		delete(list.handlers, handle.id)
		list.mu.Unlock()
	}
}

// Dispatch sends a typed event to all registered listeners immediately.
// The type parameter T is inferred from the event argument.
//
// Example:
//   Dispatch(bus, ChangeLevelEvent{MapName: "de_dust2"})
//
// Compile-time type safety ensures you cannot dispatch the wrong type.
func Dispatch[T any](bus *TypedDispatcher, event T) {
	eventType := reflect.TypeOf(event)

	if listInterface, ok := bus.listeners.Load(eventType); ok {
		list := listInterface.(*listenerList)

		list.mu.RLock()
		// Copy handlers to avoid holding lock during callbacks
		handlers := make([]EventListener[T], 0, len(list.handlers))
		for _, h := range list.handlers {
			handlers = append(handlers, h.(EventListener[T]))
		}
		list.mu.RUnlock()

		// Execute handlers without holding lock
		for _, handler := range handlers {
			handler(event)
		}
	}
}

// HasListeners returns true if there are any listeners for the given event type
func HasListeners[T any](bus *TypedDispatcher) bool {
	eventType := reflect.TypeOf((*T)(nil)).Elem()
	_, ok := bus.listeners.Load(eventType)
	return ok
}

// DispatchReflect dispatches an event using reflection instead of generics.
// This is useful when the concrete event type is not known at compile time,
// such as when dispatching through an interface{} parameter.
//
// Use this for:
// - Events received through interface{} parameters
// - Dynamic event routing where type is determined at runtime
// - Bridging non-generic code to the typed event system
//
// For better performance and type safety, prefer Dispatch[T] when the type is known.
func DispatchReflect(bus *TypedDispatcher, event interface{}) {
	eventType := reflect.TypeOf(event)

	if listInterface, ok := bus.listeners.Load(eventType); ok {
		list := listInterface.(*listenerList)

		list.mu.RLock()
		// Copy handlers to avoid holding lock during callbacks
		handlers := make([]interface{}, 0, len(list.handlers))
		for _, h := range list.handlers {
			handlers = append(handlers, h)
		}
		list.mu.RUnlock()

		// Execute handlers with proper type using reflection
		for _, h := range handlers {
			handlerValue := reflect.ValueOf(h)
			handlerValue.Call([]reflect.Value{reflect.ValueOf(event)})
		}
	}
}
