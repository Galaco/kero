package components

import (
	"github.com/galaco/kero/framework/ecs"
	"github.com/go-gl/mathgl/mgl32"
)

// Physics component for rigid body dynamics.
// Entities with this component are affected by forces and collide with other objects.
type Physics struct {
	// Linear motion
	Velocity mgl32.Vec3 // Current linear velocity (units/second)
	Force    mgl32.Vec3 // Accumulated forces this frame

	// Angular motion
	AngularVel mgl32.Vec3 // Current angular velocity (radians/second)
	Torque     mgl32.Vec3 // Accumulated torques this frame

	// Physical properties
	Mass        float32 // Mass in kg (0 = infinite mass / static object)
	Restitution float32 // Bounciness (0 = no bounce, 1 = perfect bounce)
	Friction    float32 // Surface friction (0 = ice, 1 = rubber)
	Drag        float32 // Linear drag coefficient
	AngularDrag float32 // Angular drag coefficient

	// Bullet physics integration
	// RigidBodyHandle stores the actual Bullet physics rigid body handle.
	// Using interface{} to avoid CGo dependency in component definition.
	// Physics system casts this to *collision.RigidBody when needed.
	RigidBodyHandle interface{} // Bullet rigid body handle (nil = not initialized)

	// State flags
	IsKinematic bool // If true, controlled by code (not physics simulation)
	UseGravity  bool // If true, affected by gravity
	IsSleeping  bool // If true, physics simulation is paused for this body
	IsStatic    bool // If true, never moves (optimization)
}

// IsComponent implements the ecs.Component marker interface
func (Physics) IsComponent() {}

// Register Physics component with the ECS system
func init() {
	ecs.RegisterComponent[Physics](ecs.ComponentTypePhysics)
}

// ============================================================================
// RigidBody Handle Management
// ============================================================================

// SetRigidBodyHandle stores the Bullet physics handle for this entity.
// This should be called by the physics system after creating the Bullet rigid body.
func (p *Physics) SetRigidBodyHandle(handle interface{}) {
	p.RigidBodyHandle = handle
}

// GetRigidBodyHandle retrieves the Bullet physics handle.
// Returns nil if no physics body is attached.
// Physics system should cast this to *collision.RigidBody.
func (p *Physics) GetRigidBodyHandle() interface{} {
	return p.RigidBodyHandle
}

// HasRigidBody checks if a physics body is attached to this component.
// Returns true if the handle is non-nil.
func (p *Physics) HasRigidBody() bool {
	return p.RigidBodyHandle != nil
}

// ============================================================================
// Force and Motion Management
// ============================================================================

// AddForce applies a force to the rigid body
func (p *Physics) AddForce(force mgl32.Vec3) {
	p.Force = p.Force.Add(force)
}

// AddImpulse applies an instant velocity change
func (p *Physics) AddImpulse(impulse mgl32.Vec3) {
	if p.Mass > 0 {
		p.Velocity = p.Velocity.Add(impulse.Mul(1.0 / p.Mass))
	}
}

// AddTorque applies a torque to the rigid body
func (p *Physics) AddTorque(torque mgl32.Vec3) {
	p.Torque = p.Torque.Add(torque)
}

// ClearForces resets accumulated forces and torques
func (p *Physics) ClearForces() {
	p.Force = mgl32.Vec3{0, 0, 0}
	p.Torque = mgl32.Vec3{0, 0, 0}
}

// IsMoving returns true if the body has significant velocity
func (p *Physics) IsMoving() bool {
	const threshold = 0.01
	return p.Velocity.Len() > threshold || p.AngularVel.Len() > threshold
}

// NewPhysics creates a Physics component with default values
func NewPhysics(mass float32) Physics {
	return Physics{
		Velocity:    mgl32.Vec3{0, 0, 0},
		Force:       mgl32.Vec3{0, 0, 0},
		AngularVel:  mgl32.Vec3{0, 0, 0},
		Torque:      mgl32.Vec3{0, 0, 0},
		Mass:        mass,
		Restitution: 0.5,
		Friction:       0.5,
		Drag:           0.1,
		AngularDrag:    0.1,
		RigidBodyHandle: nil,
		IsKinematic:    false,
		UseGravity:  true,
		IsSleeping:  false,
		IsStatic:    mass == 0,
	}
}

// NewStaticPhysics creates a Physics component for static (non-moving) objects
func NewStaticPhysics() Physics {
	return Physics{
		Mass:        0,
		Restitution: 0.5,
		Friction:    0.5,
		IsStatic:    true,
		UseGravity:  false,
	}
}
