package browser

import (
	"github.com/galaco/kero/framework/event"
)

// EventBusAdapter adapts Kero's event.Dispatcher to browser.EventDispatcher interface
// This allows the browser to dispatch events without depending on the concrete event system
type EventBusAdapter struct {
	dispatcher *event.Dispatcher
}

// NewEventBusAdapter creates a new adapter for Kero's event system
func NewEventBusAdapter(dispatcher *event.Dispatcher) *EventBusAdapter {
	return &EventBusAdapter{
		dispatcher: dispatcher,
	}
}

// DispatchTyped dispatches a typed event using Kero's event system
// Uses reflection-based dispatch since we receive the event as interface{}
func (a *EventBusAdapter) DispatchTyped(evt interface{}) {
	// Use reflection-based dispatch to preserve runtime type information
	// The generic DispatchTyped[T] cannot be used here because evt is interface{}
	// which would erase the concrete event type
	event.DispatchReflect(a.dispatcher.TypedDispatcher, evt)
}
