package shaders

// language=glsl
var LightMappedGenericFragment = `
    #version 410

	uniform sampler2D albedoSampler;
	uniform sampler2D lightmapSampler;

	// Flag that this material is in some way translucent
	uniform int hasTranslucentProperty;

	// Translucent variations
	uniform float alpha;
	uniform int translucent;

	// Debug Options
	uniform int renderLightmapsAsAlbedo;

	in vec2 UV;
	in vec2 LightmapUV;

    out vec4 frag_colour;

	vec4 AlbedoPass() 
	{
		if (renderLightmapsAsAlbedo == 1) {
			return texture(lightmapSampler, LightmapUV).rgba;
		}

		return texture(albedoSampler, UV).rgba;
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

	vec4 LightmapPass(in vec4 color)
	{	
		if (renderLightmapsAsAlbedo == 1) {
			return color;
		}
		if (LightmapUV.x == -1) {
			return color;
		}

		vec4 lightmapColor = vec4(texture(lightmapSampler, LightmapUV).rgb, 1.0);
		
		return color * lightmapColor;
	}

    void main() 
	{
		vec4 diffuse = AlbedoPass();
		diffuse = LightmapPass(diffuse);

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
	layout(location = 4) in vec2 vertexLightmapUV;

	out vec2 UV;
	out vec2 LightmapUV;

    void main() {
		gl_Position = projection * view * model * vec4(vertexPosition, 1.0);

    	UV = vertexUV;
    	LightmapUV = vertexLightmapUV;
    }
` + "\x00"
