package scene

import "sync"

// Manager manages the currently active scene.
// This replaces the global sceneSingleton pattern with explicit ownership.
type Manager struct {
	currentScene *StaticScene
	mu           sync.RWMutex
}

// NewManager creates a new scene manager
func NewManager() *Manager {
	return &Manager{}
}

// GetCurrentScene returns the current scene, or nil if no scene is loaded
func (m *Manager) GetCurrentScene() *StaticScene {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentScene
}

// SetCurrentScene sets the current scene
func (m *Manager) SetCurrentScene(scene *StaticScene) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentScene = scene
}

// CloseCurrentScene clears the current scene
func (m *Manager) CloseCurrentScene() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentScene = nil
}

// HasScene returns true if a scene is currently loaded
func (m *Manager) HasScene() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentScene != nil
}
