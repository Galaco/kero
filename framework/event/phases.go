package event

import (
	"reflect"
	"sync"
)

// EventPhase defines when deferred events should be processed.
// Events can be queued and processed at specific points in the frame,
// preventing cascading updates and mid-frame state corruption.
type EventPhase int

const (
	// PhaseImmediate processes events immediately (use sparingly, prefer deferred)
	PhaseImmediate EventPhase = iota

	// PhasePreUpdate runs before game logic updates (after input)
	PhasePreUpdate

	// PhasePostUpdate runs after game logic updates (before rendering)
	PhasePostUpdate

	// PhasePreRender runs before rendering begins
	PhasePreRender

	// PhasePostRender runs after frame completes
	PhasePostRender
)

// DeferredEvent stores an event to be processed during a specific phase
type DeferredEvent struct {
	Phase   EventPhase
	Type    reflect.Type
	Payload interface{}
}

// DeferredDispatcher extends TypedDispatcher with deferred event queuing.
// Events can be queued and processed at specific frame phases for predictable execution.
type DeferredDispatcher struct {
	*TypedDispatcher
	deferred map[EventPhase][]DeferredEvent
	mu       sync.RWMutex
}

// NewDeferredDispatcher creates a dispatcher with deferred event support
func NewDeferredDispatcher() *DeferredDispatcher {
	return &DeferredDispatcher{
		TypedDispatcher: NewTypedDispatcher(),
		deferred: map[EventPhase][]DeferredEvent{
			PhaseImmediate:  make([]DeferredEvent, 0),
			PhasePreUpdate:  make([]DeferredEvent, 0),
			PhasePostUpdate: make([]DeferredEvent, 0),
			PhasePreRender:  make([]DeferredEvent, 0),
			PhasePostRender: make([]DeferredEvent, 0),
		},
	}
}

// DispatchDeferred queues an event to be processed during a specific phase.
// This is the preferred way to dispatch events to avoid mid-frame state corruption.
//
// Example:
//   // Queue level change to happen after update completes
//   DispatchDeferred(bus, PhasePostUpdate, ChangeLevelEvent{MapName: "newmap"})
//
// The event will be processed when ProcessPhase(PhasePostUpdate) is called.
func DispatchDeferred[T any](bus *DeferredDispatcher, phase EventPhase, event T) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	bus.deferred[phase] = append(bus.deferred[phase], DeferredEvent{
		Phase:   phase,
		Type:    reflect.TypeOf(event),
		Payload: event,
	})
}

// ProcessPhase processes all events queued for the given phase.
// This should be called at specific points in the game loop.
//
// Typical game loop integration:
//   ProcessPhase(PhasePreUpdate)
//   input.Poll()
//   physics.Update()
//   scene.Update()
//   ProcessPhase(PhasePostUpdate)
//   ProcessPhase(PhasePreRender)
//   renderer.Render()
//   ProcessPhase(PhasePostRender)
func (bus *DeferredDispatcher) ProcessPhase(phase EventPhase) {
	bus.mu.Lock()
	events := bus.deferred[phase]
	bus.deferred[phase] = nil // Clear queue
	bus.mu.Unlock()

	// Process all events from this phase
	for _, deferredEvent := range events {
		bus.dispatchTyped(deferredEvent.Type, deferredEvent.Payload)
	}
}

// ClearPhase clears all queued events for a specific phase without processing them
func (bus *DeferredDispatcher) ClearPhase(phase EventPhase) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.deferred[phase] = nil
}

// ClearAllPhases clears all queued events across all phases
func (bus *DeferredDispatcher) ClearAllPhases() {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	for phase := range bus.deferred {
		bus.deferred[phase] = nil
	}
}

// dispatchTyped dispatches an event by type (internal helper for deferred events)
func (bus *DeferredDispatcher) dispatchTyped(eventType reflect.Type, payload interface{}) {
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
			handlerValue.Call([]reflect.Value{reflect.ValueOf(payload)})
		}
	}
}

// PendingCount returns the number of events queued for a specific phase
func (bus *DeferredDispatcher) PendingCount(phase EventPhase) int {
	bus.mu.RLock()
	defer bus.mu.RUnlock()
	return len(bus.deferred[phase])
}

// TotalPendingCount returns the total number of queued events across all phases
func (bus *DeferredDispatcher) TotalPendingCount() int {
	bus.mu.RLock()
	defer bus.mu.RUnlock()

	total := 0
	for _, events := range bus.deferred {
		total += len(events)
	}
	return total
}
