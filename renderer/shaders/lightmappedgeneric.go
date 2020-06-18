package shaders

//language=glsl
var LightMappedGenericFragment = `
    #version 410

	uniform sampler2D albedoSampler;

	uniform int alpha;

	in vec2 UV;

    out vec4 frag_colour;

	vec4 GetAlbedo(in sampler2D sampler, in vec2 uv) 
	{
		return texture(sampler, uv).rgba;
	}

    void main() 
	{
		vec4 diffuse = GetAlbedo(albedoSampler, UV);

		if (alpha == 0) {
			diffuse.a = 1;
		}

		frag_colour = diffuse;
    }
` + "\x00"

// language=glsl
var LightMappedGenericVertex = `
    #version 410

	uniform mat4 projection;
	uniform mat4 view;
	uniform mat4 model;

    layout(location = 0) in vec3 vertexPosition;
	layout(location = 1) in vec2 vertexUV;
	layout(location = 2) in vec3 vertexNormal;
	layout(location = 3) in vec4 vertexTangent;

	out vec2 UV;

    void main() {
		gl_Position = projection * view * model * vec4(vertexPosition, 1.0);

    	UV = vertexUV;
    }
` + "\x00"
