package ecs

import "reflect"

// ComponentType uniquely identifies a component type.
// Each component struct gets assigned a unique ComponentType ID.
type ComponentType uint16

const (
	ComponentTypeInvalid ComponentType = iota
	ComponentTypeTransform
	ComponentTypePhysics
	ComponentTypeModel
	ComponentTypeAnimation
	ComponentTypeBreakable
	ComponentTypeScript
	ComponentTypeLight
	ComponentTypeCamera
	ComponentTypeAudio
	ComponentTypeParticles
	// Add more component types as needed
)

// Component is a marker interface that all components must implement.
// This prevents non-component types from being used as components.
type Component interface {
	IsComponent() // Marker method - must be implemented by all components
}

// componentRegistry maps Go types to ComponentType IDs and vice versa
var componentRegistry = struct {
	typeToID map[reflect.Type]ComponentType
	idToType map[ComponentType]reflect.Type
}{
	typeToID: make(map[reflect.Type]ComponentType),
	idToType: make(map[ComponentType]reflect.Type),
}

// RegisterComponent associates a Go type with a ComponentType ID.
// This should be called in init() functions of component packages.
//
// Example:
//   func init() {
//       ecs.RegisterComponent[Transform](ecs.ComponentTypeTransform)
//   }
func RegisterComponent[T Component](componentType ComponentType) {
	var zero T
	goType := reflect.TypeOf(zero)
	componentRegistry.typeToID[goType] = componentType
	componentRegistry.idToType[componentType] = goType
}

// GetComponentType returns the ComponentType for a given Go type
func GetComponentType[T Component]() ComponentType {
	var zero T
	goType := reflect.TypeOf(zero)
	if id, exists := componentRegistry.typeToID[goType]; exists {
		return id
	}
	return ComponentTypeInvalid
}

// GetComponentGoType returns the Go type for a given ComponentType
func GetComponentGoType(componentType ComponentType) reflect.Type {
	return componentRegistry.idToType[componentType]
}
