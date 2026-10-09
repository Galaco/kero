#include <btBulletDynamicsCommon.h>
#include "Bullet-C-Api.h"
#include "bulletglue.h"

#ifdef __cplusplus
extern "C" {
#endif

plCollisionShapeHandle plNewStaticPlaneShape(plVector3 planeNormal, float planeConstant) {
	void *mem = btAlignedAlloc(sizeof(btStaticPlaneShape),16);
	return (plCollisionShapeHandle) new (mem)btStaticPlaneShape(btVector3(planeNormal[0],planeNormal[1],planeNormal[2]), planeConstant);
}

void plSetLinearVelocity(plRigidBodyHandle object, const plVector3 velocity) {
  btRigidBody* body = reinterpret_cast< btRigidBody* >(object);
  btAssert(body);
  btVector3 vel(velocity[0],velocity[1],velocity[2]);
  btTransform worldTrans = body->getWorldTransform();
  body->setLinearVelocity(vel);
  body->setWorldTransform(worldTrans);
}

void plGetLinearVelocity(plRigidBodyHandle object, plVector3 velocity) {
  btRigidBody* body = reinterpret_cast< btRigidBody* >(object);
  btAssert(body);
  btVector3 vel = body->getLinearVelocity();
  velocity[0] = vel.getX();
  velocity[1] = vel.getY();
  velocity[2] = vel.getZ();
}

void plSetGravity(plDynamicsWorldHandle world, plReal x, plReal y, plReal z) {
  btDynamicsWorld* dynamicsWorld = reinterpret_cast< btDynamicsWorld* >(world);
  dynamicsWorld->setGravity(btVector3(x,y,z));
}

void plApplyImpulse(plRigidBodyHandle object, const plVector3 impulse, const plVector3 relativePos) {
  btRigidBody* body = reinterpret_cast<btRigidBody*>(object);
  btAssert(body);
  btVector3 implse(impulse[0], impulse[1], impulse[2]);
  btVector3 relPos(relativePos[0], relativePos[1], relativePos[2]);
  body->applyImpulse(implse, relPos);
}

plCollisionShapeHandle plNewBvhTriangleMeshShape(const int* indices, int totalTriangles, const plVector3* vertices, int totalVerts)
{
	// Bullet reads the mesh for as long as the shape exists, so it gets its own copy, which plDeleteShape frees
	void* mem = btAlignedAlloc(sizeof(btTriangleMesh),16);
	btTriangleMesh* mesh = new (mem)btTriangleMesh(true, false);
	mesh->preallocateVertices(totalVerts);
	mesh->preallocateIndices(totalTriangles * 3);
	for (int i = 0; i < totalVerts; i++) {
		mesh->findOrAddVertex(btVector3(vertices[i][0], vertices[i][1], vertices[i][2]), false);
	}
	for (int i = 0; i < totalTriangles; i++) {
		mesh->addTriangleIndices(indices[i*3], indices[i*3+1], indices[i*3+2]);
	}

	mem = btAlignedAlloc(sizeof(btBvhTriangleMeshShape),16);
	return (plCollisionShapeHandle) new (mem)btBvhTriangleMeshShape(mesh, true, true);
}

void plSetActivationState(plRigidBodyHandle object, int state) {
  btRigidBody* body = reinterpret_cast<btRigidBody*>(object);
  btAssert(body);
  body->setActivationState(state);
}

void plForceActivationState(plRigidBodyHandle object, int state) {
  btRigidBody* body = reinterpret_cast<btRigidBody*>(object);
  btAssert(body);
  body->forceActivationState(state);
}

plCollisionShapeHandle plNewCapsuleShapeZ(plReal radius, plReal height) {
  void *mem = btAlignedAlloc(sizeof(btCapsuleShapeZ),16);
  return (plCollisionShapeHandle) new (mem)btCapsuleShapeZ(radius, height);
}

// Callback class for convex sweep test
struct SweepResultCallback : public btCollisionWorld::ClosestConvexResultCallback {
  SweepResultCallback(const btVector3& from, const btVector3& to)
    : btCollisionWorld::ClosestConvexResultCallback(from, to) {}
};

void plConvexSweepTest(plDynamicsWorldHandle world, plCollisionShapeHandle shape,
                       const plVector3 from, const plVector3 to, plSweepResult* result) {
  btDynamicsWorld* dynamicsWorld = reinterpret_cast<btDynamicsWorld*>(world);
  btConvexShape* convexShape = reinterpret_cast<btConvexShape*>(shape);

  btVector3 fromVec(from[0], from[1], from[2]);
  btVector3 toVec(to[0], to[1], to[2]);

  btTransform fromTrans;
  fromTrans.setIdentity();
  fromTrans.setOrigin(fromVec);

  btTransform toTrans;
  toTrans.setIdentity();
  toTrans.setOrigin(toVec);

  SweepResultCallback callback(fromVec, toVec);

  dynamicsWorld->convexSweepTest(convexShape, fromTrans, toTrans, callback);

  // Check if there was a hit (m_closestHitFraction < 1.0 means hit)
  bool hasHit = callback.m_closestHitFraction < btScalar(1.0);
  result->hasHit = hasHit ? 1 : 0;

  if (hasHit) {
    result->hitPoint[0] = callback.m_hitPointWorld.getX();
    result->hitPoint[1] = callback.m_hitPointWorld.getY();
    result->hitPoint[2] = callback.m_hitPointWorld.getZ();

    result->hitNormal[0] = callback.m_hitNormalWorld.getX();
    result->hitNormal[1] = callback.m_hitNormalWorld.getY();
    result->hitNormal[2] = callback.m_hitNormalWorld.getZ();

    result->hitFraction = callback.m_closestHitFraction;
  }
}

// Callback class for raycast
struct RayResultCallback : public btCollisionWorld::ClosestRayResultCallback {
  RayResultCallback(const btVector3& from, const btVector3& to)
    : btCollisionWorld::ClosestRayResultCallback(from, to) {}
};

void plRayTest(plDynamicsWorldHandle world, const plVector3 from, const plVector3 to,
               plRaycastResult* result) {
  btDynamicsWorld* dynamicsWorld = reinterpret_cast<btDynamicsWorld*>(world);

  btVector3 fromVec(from[0], from[1], from[2]);
  btVector3 toVec(to[0], to[1], to[2]);

  RayResultCallback callback(fromVec, toVec);

  dynamicsWorld->rayTest(fromVec, toVec, callback);

  // Check if there was a hit (m_closestHitFraction < 1.0 means hit)
  bool hasHit = callback.m_closestHitFraction < btScalar(1.0);
  result->hasHit = hasHit ? 1 : 0;

  if (hasHit) {
    result->hitPoint[0] = callback.m_hitPointWorld.getX();
    result->hitPoint[1] = callback.m_hitPointWorld.getY();
    result->hitPoint[2] = callback.m_hitPointWorld.getZ();

    result->hitNormal[0] = callback.m_hitNormalWorld.getX();
    result->hitNormal[1] = callback.m_hitNormalWorld.getY();
    result->hitNormal[2] = callback.m_hitNormalWorld.getZ();

    result->hitFraction = callback.m_closestHitFraction;
  }
}

#ifdef __cplusplus
}
#endif
