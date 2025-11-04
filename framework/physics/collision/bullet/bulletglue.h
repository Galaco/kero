#ifndef BULLETGLUE_H
#define BULLETGLUE_H

#include "Bullet-C-Api.h"

#ifdef __cplusplus
extern "C" {
#endif

plCollisionShapeHandle plNewStaticPlaneShape(plVector3 planeNormal, float planeConstant);
void plSetLinearVelocity(plRigidBodyHandle object, const plVector3 velocity);
void plGetLinearVelocity(plRigidBodyHandle object, plVector3 velocity);
void plSetGravity(plDynamicsWorldHandle world, plReal x, plReal y, plReal z);
void plApplyImpulse(plRigidBodyHandle object, const plVector3 impulse, const plVector3 relativePos);
plCollisionShapeHandle btNewBvhTriangleIndexVertexArray(int* indices, plVector3* vertices, int totalTriangles, int totalVerts);
plCollisionShapeHandle btNewBvhTriangleMeshShape(plCollisionShapeHandle indexVertexArrays);
void plSetActivationState(plRigidBodyHandle object, int state);
void plForceActivationState(plRigidBodyHandle object, int state);
plCollisionShapeHandle plNewCapsuleShapeZ(plReal radius, plReal height);

// Sweep test result structure
typedef struct {
    int hasHit;
    plVector3 hitPoint;
    plVector3 hitNormal;
    plReal hitFraction;
} plSweepResult;

// Raycast result structure
typedef struct plRaycastResult {
    int hasHit;
    plVector3 hitPoint;
    plVector3 hitNormal;
    plReal hitFraction;
} plRaycastResult;

// Sweep a convex shape from one position to another
void plConvexSweepTest(plDynamicsWorldHandle world, plCollisionShapeHandle shape,
                       const plVector3 from, const plVector3 to, plSweepResult* result);

// Raycast test
void plRayTest(plDynamicsWorldHandle world, const plVector3 from, const plVector3 to,
               plRaycastResult* result);

#ifdef __cplusplus
}
#endif

#endif

