package engine

import (
	"github.com/galaco/kero/framework/ecs"
	"github.com/galaco/kero/framework/entity"
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/framework/metrics"
	frameworkScene "github.com/galaco/kero/framework/scene"
	"github.com/galaco/kero/framework/window"
	"github.com/galaco/kero/gui"
	"github.com/galaco/kero/middleware"
	"github.com/galaco/kero/physics"
	"github.com/galaco/kero/renderer"
	"github.com/galaco/kero/scene"
)

// Engine provides a central context that owns all game systems and their dependencies.
// This replaces the previous global singleton pattern with explicit dependency injection.
type Engine struct {
	// Core infrastructure
	eventBus       *event.Dispatcher
	fileSystem     filesystem.FileSystem
	entityRegistry *entity.Registry
	metrics        *metrics.Collector

	// Game state
	sceneManager *frameworkScene.Manager
	ecsWorld     *ecs.World // Phase 4: Pure ECS - no legacy bridge needed

	// Systems (in dependency order)
	input    *middleware.Input
	physics  *physics.PhysicsSystem
	scene    *scene.Scene
	renderer *renderer.Renderer
	gui      *gui.Gui

	// Window (owned by engine for lifecycle management)
	window *window.Window
}

// NewEngine creates a new Engine instance with all systems uninitialized.
// Call Initialize() after setting up all dependencies.
func NewEngine() *Engine {
	return &Engine{}
}

// EventBus returns the event dispatcher
func (e *Engine) EventBus() *event.Dispatcher {
	return e.eventBus
}

// SetEventBus sets the event dispatcher
func (e *Engine) SetEventBus(eventBus *event.Dispatcher) {
	e.eventBus = eventBus
}

// FileSystem returns the filesystem
func (e *Engine) FileSystem() filesystem.FileSystem {
	return e.fileSystem
}

// SetFileSystem sets the filesystem
func (e *Engine) SetFileSystem(fs filesystem.FileSystem) {
	e.fileSystem = fs
}

// EntityRegistry returns the entity class registry
func (e *Engine) EntityRegistry() *entity.Registry {
	return e.entityRegistry
}

// SetEntityRegistry sets the entity class registry
func (e *Engine) SetEntityRegistry(registry *entity.Registry) {
	e.entityRegistry = registry
}

// SceneManager returns the scene manager
func (e *Engine) SceneManager() *frameworkScene.Manager {
	return e.sceneManager
}

// SetSceneManager sets the scene manager
func (e *Engine) SetSceneManager(manager *frameworkScene.Manager) {
	e.sceneManager = manager
}

// Input returns the input middleware
func (e *Engine) Input() *middleware.Input {
	return e.input
}

// SetInput sets the input middleware
func (e *Engine) SetInput(input *middleware.Input) {
	e.input = input
}

// Physics returns the physics system
func (e *Engine) Physics() *physics.PhysicsSystem {
	return e.physics
}

// SetPhysics sets the physics system
func (e *Engine) SetPhysics(physics *physics.PhysicsSystem) {
	e.physics = physics
}

// Scene returns the scene system
func (e *Engine) Scene() *scene.Scene {
	return e.scene
}

// SetScene sets the scene system
func (e *Engine) SetScene(scene *scene.Scene) {
	e.scene = scene
}

// Renderer returns the renderer
func (e *Engine) Renderer() *renderer.Renderer {
	return e.renderer
}

// SetRenderer sets the renderer
func (e *Engine) SetRenderer(renderer *renderer.Renderer) {
	e.renderer = renderer
}

// GUI returns the GUI system
func (e *Engine) GUI() *gui.Gui {
	return e.gui
}

// SetGUI sets the GUI system
func (e *Engine) SetGUI(gui *gui.Gui) {
	e.gui = gui
}

// Window returns the window
func (e *Engine) Window() *window.Window {
	return e.window
}

// SetWindow sets the window
func (e *Engine) SetWindow(w *window.Window) {
	e.window = w
}

// ECSWorld returns the ECS world
func (e *Engine) ECSWorld() *ecs.World {
	return e.ecsWorld
}

// SetECSWorld sets the ECS world
func (e *Engine) SetECSWorld(world *ecs.World) {
	e.ecsWorld = world
}

// Metrics returns the metrics collector
func (e *Engine) Metrics() *metrics.Collector {
	return e.metrics
}

// SetMetrics sets the metrics collector
func (e *Engine) SetMetrics(metrics *metrics.Collector) {
	e.metrics = metrics
}
