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
	// Check if we're starting on ground (for StayOnGround logic)
	wasOnGround := cc.CheckGround(currentPos)

	if desiredMove.Len() < 0.001 {
		// No movement desired
		return MoveResult{
			FinalPosition: currentPos,
			OnGround:      wasOnGround,
			HitWall:       false,
		}
	}

	targetPos := currentPos.Add(desiredMove)

	// First sweep test along desired path
	result := bullet.BulletConvexSweepTest(cc.world, cc.capsuleShape, currentPos, targetPos)

	// Check if we hit something
	if !result.HasHit {
		// No collision, move freely
		// Apply StayOnGround to snap down to slopes
		finalPos := cc.StayOnGround(targetPos, currentPos, wasOnGround)
		return MoveResult{
			FinalPosition: finalPos,
			OnGround:      cc.CheckGround(finalPos),
			HitWall:       false,
		}
	}

	// Debug: Log collision details
	fmt.Printf("Sweep: hasHit=%v, fraction=%.3f, normal=(%.2f,%.2f,%.2f)\n",
		result.HasHit, result.HitFraction, result.HitNormal[0], result.HitNormal[1], result.HitNormal[2])

	// Check if we're moving into the surface or away from it
	// If moving away or parallel (dot >= 0), we can ignore this collision (it's the floor beneath us)
	// If moving into it (dot < 0), we need to handle the collision (it's a wall or upward slope)
	moveDir := desiredMove.Normalize()
	dotProduct := moveDir.Dot(result.HitNormal)

	// If hit fraction is very small, normal is pointing up, AND we're not moving into the surface
	// then we're hitting the floor we're standing on - treat as no collision for horizontal movement
	if result.HitFraction < 0.01 && result.HitNormal[2] > 0.7 && dotProduct >= -0.01 {
		// Standing on ground, moving parallel/away from surface - no wall collision
		// Apply StayOnGround to snap down to slopes
		finalPos := cc.StayOnGround(targetPos, currentPos, wasOnGround)
		return MoveResult{
			FinalPosition: finalPos,
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

	// Apply StayOnGround after sliding (if we're still on ground and didn't hit a wall)
	if !slideResult.HitWall && slideResult.OnGround {
		slideResult.FinalPosition = cc.StayOnGround(slideResult.FinalPosition, currentPos, wasOnGround)
	}

	return slideResult
}

// slideMove slides the character along a collision surface
func (cc *CharacterController) slideMove(currentPos, desiredMove mgl32.Vec3, firstHit bullet.SweepResult) MoveResult {
	// Move to just before hit point (with tiny safety margin)
	safetyMargin := float32(0.01)

	// Calculate how far we can move before hitting
	var hitPos mgl32.Vec3
	var remainingMove mgl32.Vec3

	if firstHit.HitFraction > safetyMargin {
		// We have some distance before the collision
		hitPos = currentPos.Add(desiredMove.Mul(firstHit.HitFraction - safetyMargin))
		remainingMove = desiredMove.Mul(1.0 - firstHit.HitFraction)
	} else {
		// Already at/very close to collision point - use full desired move for projection
		// This handles the case where we're already touching a surface (like base of slope)
		hitPos = currentPos
		remainingMove = desiredMove
	}

	// Calculate slide direction (project remaining movement onto surface)
	dotProduct := remainingMove.Dot(firstHit.HitNormal)
	slideDir := remainingMove.Sub(firstHit.HitNormal.Mul(dotProduct))

	// Determine if this is a wall (steep surface) or slope (walkable)
	isWall := firstHit.HitNormal[2] <= 0.7

	// If slide direction is too small, we can't make meaningful slide progress
	// Try stepping up instead (Source Engine's "high road" approach)
	// This handles tiny geometry seams/artifacts on slopes
	if slideDir.Len() < 0.0001 {
		// Try stepping up as fallback (even on slopes)
		stepResult := cc.tryStepUp(currentPos, desiredMove, firstHit)
		if stepResult.FinalPosition != currentPos {
			// Step up succeeded
			return stepResult
		}
		// Both slide and step failed - no movement possible
		return MoveResult{
			FinalPosition: hitPos,
			OnGround:      cc.CheckGround(hitPos),
			HitWall:       isWall,
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

	// Determine if we hit a wall (steep surface) or a slope (walkable surface)
	// A slope has normal Z > 0.7 (less than ~45 degrees)
	// A wall has normal Z <= 0.7 (steep/vertical)
	isWall = firstHit.HitNormal[2] <= 0.7

	return MoveResult{
		FinalPosition: finalPos,
		OnGround:      cc.CheckGround(finalPos),
		HitWall:       isWall, // Only true for actual walls, not slopes
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

// StayOnGround attempts to keep the player attached to downward slopes.
// This is called after horizontal movement to snap the player down to ground.
// Returns the adjusted position if ground was found, otherwise returns the input position.
// startPos is the position before movement (to detect if we moved upward).
func (cc *CharacterController) StayOnGround(position mgl32.Vec3, startPos mgl32.Vec3, wasOnGround bool) mgl32.Vec3 {
	// Only snap down if we were on ground (prevents snapping during jumps/falling)
	if !wasOnGround {
		return position
	}

	// Don't snap down if we moved upward (climbing a slope)
	// This prevents fighting against upward slope movement
	if position.Z() > startPos.Z()+0.1 {
		return position
	}

	// Trace down from current position by stepHeight distance (18 units in Source Engine)
	// This matches Source Engine's StayOnGround behavior
	downVector := mgl32.Vec3{0, 0, -float32(cc.stepHeight)}
	downTarget := position.Add(downVector)

	// Sweep test downward
	result := bullet.BulletConvexSweepTest(cc.world, cc.capsuleShape, position, downTarget)

	if !result.HasHit {
		// No ground found within stepHeight - we're airborne
		return position
	}

	// Check if surface is walkable (not too steep)
	upDot := result.HitNormal.Dot(mgl32.Vec3{0, 0, 1})
	if upDot <= 0.7 {
		// Too steep, don't snap down
		return position
	}

	// Ground found - snap player down to maintain contact
	// Move down by the hit fraction distance
	snapDistance := float32(cc.stepHeight) * result.HitFraction
	snappedPos := position.Add(mgl32.Vec3{0, 0, -snapDistance})

	return snappedPos
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
