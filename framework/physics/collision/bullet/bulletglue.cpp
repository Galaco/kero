#include <btBulletDynamicsCommon.h>
#include "Bullet-C-Api.h"

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

#ifdef __cplusplus
}
#endif
