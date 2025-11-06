package shaders

//language=glsl
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

// LightMappedGenericInstancedVertex is the instanced version of the vertex shader
// Supports per-instance transforms and fade distances
//language=glsl
var LightMappedGenericInstancedVertex = `
    #version 410

	uniform mat4 projection;
	uniform mat4 view;
	uniform vec3 cameraPosition;  // For fade distance calculation

    layout(location = 0) in vec3 vertexPosition;
	layout(location = 1) in vec2 vertexUV;
	layout(location = 2) in vec3 vertexNormal;
	layout(location = 3) in vec4 vertexTangent;
	layout(location = 4) in vec2 vertexLightmapUV;

	// Instance attributes (per-instance data)
	// Note: mat4 constructor treats vec4s as COLUMNS, not rows
	layout(location = 5) in vec4 instanceModelMatrixCol0;
	layout(location = 6) in vec4 instanceModelMatrixCol1;
	layout(location = 7) in vec4 instanceModelMatrixCol2;
	layout(location = 8) in vec4 instanceModelMatrixCol3;
	layout(location = 9) in vec2 instanceFadeMinMax;  // x=fadeMin, y=fadeMax

	out vec2 UV;
	out vec2 LightmapUV;
	out float fadeAlpha;

    void main() {
		// Reconstruct model matrix from instance attributes
		// mat4() constructor takes COLUMNS, and we're storing in column-major order
		mat4 instanceModelMatrix = mat4(
			instanceModelMatrixCol0,
			instanceModelMatrixCol1,
			instanceModelMatrixCol2,
			instanceModelMatrixCol3
		);

		gl_Position = projection * view * instanceModelMatrix * vec4(vertexPosition, 1.0);

    	UV = vertexUV;
    	LightmapUV = vertexLightmapUV;

		// Calculate fade alpha based on distance
		if (instanceFadeMinMax.y > 0.0) {  // fadeMax > 0 means fading enabled
			// Extract world position from model matrix (translation in last column)
			vec3 worldPos = vec3(instanceModelMatrix[3][0],
								instanceModelMatrix[3][1],
								instanceModelMatrix[3][2]);

			float dist = distance(worldPos, cameraPosition);
			float fadeMin = instanceFadeMinMax.x;
			float fadeMax = instanceFadeMinMax.y;

			// Calculate fade (1.0 = opaque, 0.0 = transparent)
			// Clamp to [0,1] range
			fadeAlpha = 1.0 - clamp((dist - fadeMin) / (fadeMax - fadeMin), 0.0, 1.0);
		} else {
			fadeAlpha = 1.0;  // No fading
		}
    }
` + "\x00"

// LightMappedGenericInstancedFragment is the instanced version of the fragment shader
// Applies per-instance fade alpha
//language=glsl
var LightMappedGenericInstancedFragment = `
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
	in float fadeAlpha;  // From vertex shader

    out vec4 frag_colour;

	vec4 AlbedoPass()
	{
		if (renderLightmapsAsAlbedo == 1) {
			return texture(lightmapSampler, LightmapUV).rgba;
		}

		return texture(albedoSampler, UV).rgba;
	}

	// Handle transparency rules here
	vec4 AlphaPass(in vec4 color)
	{
		if (hasTranslucentProperty == 0) {
			// Ignore material alpha channel
			color.a = 1;
			return color;
		}

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

		// Apply distance-based fade alpha
		diffuse.a *= fadeAlpha;

		// Discard fully transparent fragments (optimization)
		if (diffuse.a < 0.01) {
			discard;
		}

		frag_colour = diffuse;
    }
` + "\x00"
