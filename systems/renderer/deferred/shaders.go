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

	out vec2 UV;
	out vec3 Normal;
	out vec3 Position;
	
	void main() {
		gl_Position = projection * view * model * vec4(vertexPosition, 1.0);

		UV = vertexUV;
		Normal = (model * vec4(vertexNormal, 0.0)).xyz;
		Position = (model * vec4(vertexPosition, 1.0)).xyz;
	}
` + "\x00"

// language=glsl
var GeometryPassFragment = `
	#version 410
	
	uniform sampler2D albedoSampler;
	
	in vec2 UV;
	in vec3 Normal;
	in vec3 Position;
	
	layout(location = 0) out vec4 DiffuseOut;
	layout(location = 1) out vec4 NormalOut;
	layout(location = 2) out vec4 PositionOut;
	//layout(location = 3) out vec3 UVOut;

	vec4 GetAlbedo(in sampler2D sampler, in vec2 uv) 
	{
		return texture(sampler, uv).rgba;
	}
	
	void main() {
		DiffuseOut = GetAlbedo(albedoSampler, UV);
		NormalOut = vec4(normalize(Normal), 1);
		PositionOut = vec4(Position, 1);
		//UVOut = vec3(UV, 0.0);
	}
` + "\x00"



// language=glsl
var DirectionalLightPassVertex = `
	#version 410
	out vec2 fsUv;

	// full screen triangle vertices.
	const vec2 verts[3] = vec2[](vec2(-1, -1), vec2(3, -1), vec2(-1, 3));
	const vec2 uvs[3] = vec2[](vec2(0, 0), vec2(2, 0), vec2(0, 2));
	
	void main() {
		fsUv = uvs[gl_VertexID];
		gl_Position = vec4(verts[gl_VertexID], 0.0, 1.0);
	}
` + "\x00"

// language=glsl
var DirectionalLightPassFragment = `
	#version 410
	
	in vec2 fsUv;

	uniform sampler2D uColorTex;
	uniform sampler2D uNormalTex;
	uniform sampler2D uPositionTex;

	out vec4 outColor;
	
	void main() {
		outColor = vec4(texture(uColorTex, fsUv).xyz, 1.0);
	}
` + "\x00"
