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

	out vec3 Position;
	out vec3 Normal;
	out vec2 UV;
	
	void main() {
		gl_Position = projection * view * model * vec4(vertexPosition, 1.0);

		Position = (model * vec4(vertexPosition, 1.0)).xyz;
		Normal = (model * vec4(vertexNormal, 0.0)).xyz;
		UV = vertexUV;
	}
` + "\x00"

// language=glsl
var GeometryPassFragment = `
	#version 410
	
	uniform sampler2D albedoSampler;
	
	in vec3 Position;
	in vec3 Normal;
	in vec2 UV;
	
	layout (location = 0) out vec3 PositionOut;
	layout (location = 1) out vec3 NormalOut;
	layout (location = 2) out vec4 AlbedoSpecularOut;

	vec3 GetAlbedo(in sampler2D sampler, in vec2 uv) 
	{
		return texture(sampler, uv).rgb;
	}

	float GetSpecular(in sampler2D sampler, in vec2 uv) 
	{
		return 1;
		// return texture(sampler, uv).r;
	}
	
	void main() {
		PositionOut = Position;
		NormalOut = normalize(Normal);
		AlbedoSpecularOut.rgb = GetAlbedo(albedoSampler, UV);
		AlbedoSpecularOut.a = GetSpecular(albedoSampler, UV);
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

	uniform sampler2D uPositionTex;
	uniform sampler2D uNormalTex;
	uniform sampler2D uColorTex;

	out vec4 outColor;
	
	void main() {
		outColor = vec4(texture(uColorTex, fsUv).xyz, 1.0);
	}
` + "\x00"
