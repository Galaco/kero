package event

// Type represents the event name.
// This could be derived via reflection, at a substantial
// drop in performance.
type Type string

// Dispatchable Generic event manager message interface
// All messages need to implement this
type Dispatchable interface {
	// Type
	Type() Type
}

type Receiveable func(interface{})
