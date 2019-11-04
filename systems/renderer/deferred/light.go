package deferred

import "github.com/go-gl/mathgl/mgl32"

type BaseLight struct {
	Color            mgl32.Vec3
	DiffuseIntensity float32
}

type DirectionalLight struct {
	BaseLight
	AmbientColor     mgl32.Vec3
	AmbientIntensity float32
	Direction        mgl32.Vec3
}

type Attenuation struct {
	Constant, Linear, Exponential float32
}

type PointLight struct {
	Position    mgl32.Vec3
	Attenuation Attenuation
}
