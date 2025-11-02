package views

// View represents a renderable GUI component
type View interface {
	// Render draws the view with the given delta time
	Render(dt float32)
}

// UpdatableView represents a view that needs per-frame updates
type UpdatableView interface {
	View
	// Update performs per-frame logic updates
	Update(dt float32)
}
