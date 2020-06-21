package shaders

//language=glsl
var LightMappedGenericFragment = `
    #version 410

	uniform sampler2D albedoSampler;

	// Flag that this material is in some way translucent
	uniform int hasTranslucentProperty;

	// Translucent variations
	uniform float alpha;
	uniform int translucent;

	in vec2 UV;

    out vec4 frag_colour;

	vec4 GetAlbedo(in sampler2D sampler, in vec2 uv) 
	{
		return texture(sampler, uv).rgba;
	}


	// Handle transparency rules here
	// @TODO review various alpha affecting rules priority
	vec4 AlphaPass(in vec4 color)
	{	
		if (hasTranslucentProperty == 0) {
			// Ignore material alpha channel
			color.a = 1;
			return color;
		}
		// The $translucent property just means use texture alpha channel. i.e 0 processing if enabled

		// $alpha property applies a single alpha value across the entire texture 
		if (alpha != 0) {
			color.a = alpha;
		}

		return color;
	}

    void main() 
	{
		vec4 diffuse = GetAlbedo(albedoSampler, UV);

		diffuse = AlphaPass(diffuse);

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
