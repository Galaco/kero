package deferred

// language=glsl
var GeometryPassVertex = `
	#version 410
	
	uniform mat4 projection;
	uniform mat4 view;
	uniform mat4 model;

	layout(location = 0) in vec3 vertexPosition;
	layout(location = 1) in vec2 vertexUV;
	layout(location = 2) in vec3 vertexNormal;
	layout(location = 3) in vec4 vertexTangent;

	out vec2 TexCoord0;
	out vec3 Normal0;
	out vec3 WorldPos0;
	
	void main() {
		gl_Position = projection * view * model * vec4(vertexPosition, 1.0);
		TexCoord0 = vertexUV;
		Normal0 = (model * vec4(vertexNormal, 0.0)).xyz;
		WorldPos0 = (model * vec4(vertexPosition, 1.0)).xyz;
	}
` + "\x00"

// language=glsl
var GeometryPassFragment = `
	#version 410
	
	uniform sampler2D albedoSampler;
	
	in vec2 TexCoord0;
	in vec3 Normal0;
	in vec3 WorldPos0;
	
	layout(location = 0) out vec3 WorldPosOut;
	layout(location = 1) out vec3 DiffuseOut;
	layout(location = 2) out vec3 NormalOut;
	layout(location = 3) out vec3 TexCoordOut;
	
	void main() {
		WorldPosOut = WorldPos0;
		DiffuseOut = texture(albedoSampler, TexCoord0).xyz;
		NormalOut = normalize(Normal0);
		TexCoordOut = vec3(TexCoord0, 0.0);
	}
` + "\x00"
