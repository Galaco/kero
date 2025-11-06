package shaders

// WorldVertexTransitionVertex is the vertex shader for 2-texture displacement blending
// language=glsl
var WorldVertexTransitionVertex = `
    #version 410

	uniform mat4 projection;
	uniform mat4 view;
	uniform mat4 model;

    layout(location = 0) in vec3 vertexPosition;
	layout(location = 1) in vec2 vertexUV;
	layout(location = 2) in vec3 vertexNormal;
	layout(location = 3) in vec4 vertexTangent;
	layout(location = 4) in vec2 vertexLightmapUV;
	layout(location = 5) in float vertexBlendWeight;  // Blend weight for texture 2 (0.0-1.0)

	out vec2 UV;
	out vec2 LightmapUV;
	out float BlendWeight;

    void main() {
		gl_Position = projection * view * model * vec4(vertexPosition, 1.0);

    	UV = vertexUV;
    	LightmapUV = vertexLightmapUV;
		BlendWeight = vertexBlendWeight;
    }
` + "\x00"

// WorldVertexTransitionFragment is the fragment shader for 2-texture displacement blending
// language=glsl
var WorldVertexTransitionFragment = `
    #version 410

	uniform sampler2D basetextureSampler;
	uniform sampler2D basetexture2Sampler;
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
	in float BlendWeight;

    out vec4 frag_colour;

	vec4 AlbedoPass()
	{
		if (renderLightmapsAsAlbedo == 1) {
			return texture(lightmapSampler, LightmapUV).rgba;
		}

		// Sample both textures
		vec4 tex1 = texture(basetextureSampler, UV);
		vec4 tex2 = texture(basetexture2Sampler, UV);


		// Blend between them based on vertex weight
		// BlendWeight: 0.0 = full tex1, 1.0 = full tex2
		return mix(tex1, tex2, BlendWeight);
	}

	// Handle transparency rules here
	vec4 AlphaPass(in vec4 color)
	{
		if (hasTranslucentProperty == 0) {
			// Ignore material alpha channel
			color.a = 1;
			return color;
		}
		// The $translucent property just means use texture alpha channel

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
