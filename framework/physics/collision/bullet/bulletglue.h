#ifndef BULLETGLUE_H
#define BULLETGLUE_H

#include "Bullet-C-Api.h"

plCollisionShapeHandle plNewStaticPlaneShape(plVector3 planeNormal, float planeConstant);
void plSetLinearVelocity(plRigidBodyHandle object, const plVector3 velocity);
void plGetLinearVelocity(plRigidBodyHandle object, plVector3 velocity);
void plSetGravity(plDynamicsWorldHandle world, plReal x, plReal y, plReal z);
void plApplyImpulse(plRigidBodyHandle object, const plVector3 impulse, const plVector3 relativePos);
plCollisionShapeHandle btNewBvhTriangleIndexVertexArray(int* indices, plVector3* vertices, int totalTriangles, int totalVerts);
plCollisionShapeHandle btNewBvhTriangleMeshShape(plCollisionShapeHandle indexVertexArrays);

#endif

