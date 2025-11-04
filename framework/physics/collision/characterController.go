package collision

import (
	"fmt"
	"math"

	"github.com/galaco/kero/framework/physics/collision/bullet"
	"github.com/go-gl/mathgl/mgl32"
)

// CharacterController handles player physics collision using sweep tests
type CharacterController struct {
	capsuleShape bullet.BulletCollisionShapeHandle
	world        bullet.BulletDynamicWorldHandle
	height       float64
	radius       float64
	stepHeight   float64
}

// NewCharacterController creates a new character controller
func NewCharacterController(world bullet.BulletDynamicWorldHandle, capsuleShape bullet.BulletCollisionShapeHandle, height, radius, stepHeight float64) *CharacterController {
	return &CharacterController{
		capsuleShape: capsuleShape,
		world:        world,
		height:       height,
		radius:       radius,
		stepHeight:   stepHeight,
	}
}

// NewCharacterControllerFromInterfaces creates a character controller from interface{} parameters.
// This is used to avoid import cycles when the player package needs to create a controller
// but can't import the bullet package directly.
func NewCharacterControllerFromInterfaces(world interface{}, capsuleShape interface{}, height, radius, stepHeight float64) *CharacterController {
	// Type assert the world handle
	worldHandle, ok := world.(bullet.BulletDynamicWorldHandle)
	if !ok {
		// Type assertion failed - return nil
		return nil
	}

	// Type assert the capsule shape handle
	shapeHandle, ok := capsuleShape.(bullet.BulletCollisionShapeHandle)
	if !ok {
		// Type assertion failed - return nil
		return nil
	}

	return NewCharacterController(worldHandle, shapeHandle, height, radius, stepHeight)
}

// MoveResult contains the result of a move attempt
type MoveResult struct {
	FinalPosition mgl32.Vec3
	OnGround      bool
	HitWall       bool
}

// Move attempts to move the character from currentPos by desiredMove,
// handling collisions with sliding and step climbing
func (cc *CharacterController) Move(currentPos, desiredMove mgl32.Vec3) MoveResult {
	if desiredMove.Len() < 0.001 {
		// No movement desired
		return MoveResult{
			FinalPosition: currentPos,
			OnGround:      cc.CheckGround(currentPos),
			HitWall:       false,
		}
	}

	targetPos := currentPos.Add(desiredMove)

	// First sweep test along desired path
	result := bullet.BulletConvexSweepTest(cc.world, cc.capsuleShape, currentPos, targetPos)

	// Check if we hit something
	if !result.HasHit {
		// No collision, move freely
		return MoveResult{
			FinalPosition: targetPos,
			OnGround:      cc.CheckGround(targetPos),
			HitWall:       false,
		}
	}

	// Debug: Log collision details
	fmt.Printf("Sweep: hasHit=%v, fraction=%.3f, normal=(%.2f,%.2f,%.2f)\n",
		result.HasHit, result.HitFraction, result.HitNormal[0], result.HitNormal[1], result.HitNormal[2])

	// If hit fraction is very small and normal is pointing up, we're hitting the floor we're standing on
	// This is expected contact - treat it as no collision for horizontal movement
	if result.HitFraction < 0.01 && result.HitNormal[2] > 0.7 {
		// Standing on ground, no wall collision
		return MoveResult{
			FinalPosition: targetPos,
			OnGround:      true,
			HitWall:       false,
		}
	}

	// Hit something! Check if it's a step we can climb
	if cc.canStepUp(result.HitNormal) {
		stepResult := cc.tryStepUp(currentPos, desiredMove, result)
		if stepResult.FinalPosition != currentPos {
			// Successfully stepped up
			return stepResult
		}
	}

	// Can't step up, slide along surface
	slideResult := cc.slideMove(currentPos, desiredMove, result)
	return slideResult
}

// slideMove slides the character along a collision surface
func (cc *CharacterController) slideMove(currentPos, desiredMove mgl32.Vec3, firstHit bullet.SweepResult) MoveResult {
	// Move to just before hit point (with tiny safety margin)
	safetyMargin := float32(0.01)
	hitPos := currentPos.Add(desiredMove.Mul(firstHit.HitFraction - safetyMargin))
	if firstHit.HitFraction < safetyMargin {
		// Already at collision point
		return MoveResult{
			FinalPosition: currentPos,
			OnGround:      cc.CheckGround(currentPos),
			HitWall:       true,
		}
	}

	// Calculate slide direction (project remaining movement onto surface)
	remainingMove := desiredMove.Mul(1.0 - firstHit.HitFraction)
	dotProduct := remainingMove.Dot(firstHit.HitNormal)
	slideDir := remainingMove.Sub(firstHit.HitNormal.Mul(dotProduct))

	// Clamp slide to prevent climbing steep surfaces
	if slideDir.Len() < 0.001 {
		return MoveResult{
			FinalPosition: hitPos,
			OnGround:      cc.CheckGround(hitPos),
			HitWall:       true,
		}
	}

	// Try sliding movement
	slideTarget := hitPos.Add(slideDir)
	slideResult := bullet.BulletConvexSweepTest(cc.world, cc.capsuleShape, hitPos, slideTarget)

	finalPos := hitPos
	if slideResult.HasHit {
		// Hit again during slide, stop at second hit point
		finalPos = hitPos.Add(slideDir.Mul(slideResult.HitFraction))
	} else {
		// Slide succeeded
		finalPos = slideTarget
	}

	return MoveResult{
		FinalPosition: finalPos,
		OnGround:      cc.CheckGround(finalPos),
		HitWall:       true,
	}
}

// canStepUp checks if a collision normal indicates a step (horizontal surface)
func (cc *CharacterController) canStepUp(hitNormal mgl32.Vec3) bool {
	// Only step up for mostly horizontal collisions (not floor or ceiling)
	verticalDot := math.Abs(float64(hitNormal.Z()))
	return verticalDot < 0.7 // Less than ~45 degrees from horizontal
}

// tryStepUp attempts to climb a step
func (cc *CharacterController) tryStepUp(currentPos, desiredMove mgl32.Vec3, blockHit bullet.SweepResult) MoveResult {
	// Try moving up by step height
	upVector := mgl32.Vec3{0, 0, float32(cc.stepHeight)}
	upPos := currentPos.Add(upVector)

	// Check if there's space above to fit the capsule
	upSweep := bullet.BulletConvexSweepTest(cc.world, cc.capsuleShape, currentPos, upPos)
	if upSweep.HasHit {
		// Can't fit above (ceiling or overhang)
		return MoveResult{FinalPosition: currentPos}
	}

	// Try moving forward at elevated position
	forwardTarget := upPos.Add(desiredMove)
	forwardSweep := bullet.BulletConvexSweepTest(cc.world, cc.capsuleShape, upPos, forwardTarget)

	var finalPos mgl32.Vec3
	if forwardSweep.HasHit {
		// Hit something while moving forward, stop at hit point
		finalPos = upPos.Add(desiredMove.Mul(forwardSweep.HitFraction))
	} else {
		// Forward movement succeeded
		finalPos = forwardTarget
	}

	// Try to drop down to ground level (snap to floor after stepping)
	downTarget := finalPos.Sub(upVector)
	downSweep := bullet.BulletConvexSweepTest(cc.world, cc.capsuleShape, finalPos, downTarget)

	if downSweep.HasHit {
		// Found ground, snap to it
		finalPos = finalPos.Sub(upVector.Mul(downSweep.HitFraction))
	}

	// Only accept step if we actually moved forward
	horizontalMove := mgl32.Vec2{finalPos.X() - currentPos.X(), finalPos.Y() - currentPos.Y()}
	if horizontalMove.Len() > 0.01 {
		return MoveResult{
			FinalPosition: finalPos,
			OnGround:      true, // Just stepped, assume on ground
			HitWall:       false,
		}
	}

	// Step didn't help, return original position
	return MoveResult{FinalPosition: currentPos}
}

// CheckGround performs a raycast down to check if character is on ground
func (cc *CharacterController) CheckGround(position mgl32.Vec3) bool {
	// Raycast from slightly below capsule center down
	// Use a small offset to start inside the capsule bottom
	rayStart := position.Sub(mgl32.Vec3{0, 0, float32(cc.height/2 - cc.radius)})
	// Check stepHeight units down (18 units in Source Engine) to stay attached to slopes
	// This matches Source Engine's StayOnGround behavior
	rayEnd := rayStart.Sub(mgl32.Vec3{0, 0, float32(cc.stepHeight)})

	result := bullet.BulletRayTest(cc.world, rayStart, rayEnd)

	if !result.HasHit {
		return false
	}

	// Check if the surface is walkable (not too steep)
	// dot with up vector should be > 0.7 (less than ~45 degree slope)
	upDot := result.HitNormal.Dot(mgl32.Vec3{0, 0, 1})
	isWalkable := upDot > 0.7

	if !isWalkable {
		return false
	}

	// Check distance to ground (within stepHeight range)
	distToGround := rayStart.Z() - result.HitPoint.Z()
	return distToGround < float32(cc.stepHeight)
}

// GetWorld returns the physics world handle
func (cc *CharacterController) GetWorld() bullet.BulletDynamicWorldHandle {
	return cc.world
}

// GetCapsuleShape returns the capsule collision shape
func (cc *CharacterController) GetCapsuleShape() bullet.BulletCollisionShapeHandle {
	return cc.capsuleShape
}
