package shaders

//language=glsl
var LightMappedGenericFragment = `
    #version 410

	uniform sampler2D albedoSampler;
	uniform sampler2D lightmapSampler;
	// Lightmaps are stored at half brightness, so that light can be up to twice as bright as a surface's texture
	const float lightmapScale = 2.0;
	uniform sampler2D bumpmapSampler;

	// Flag that this material is in some way translucent
	uniform int hasTranslucentProperty;

	// Translucent variations
	uniform float alpha;
	uniform int translucent;

	// $alphatest discards the parts of the base texture whose alpha is below the reference
	uniform int alphaTest;
	uniform float alphaTestReference;

	// Debug Options
	uniform int renderLightmapsAsAlbedo;

	in vec2 UV;
	in vec2 LightmapUV;

    out vec4 frag_colour;

	vec4 AlbedoPass() 
	{
		if (renderLightmapsAsAlbedo == 1) {
			return vec4(texture(lightmapSampler, LightmapUV).rgb * lightmapScale, 1.0);
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

		// Only $translucent uses the texture's alpha channel
		if (translucent == 0) {
			color.a = 1;
		}

		// $alpha fades the whole material. 0 means it is not set.
		if (alpha != 0) {
			color.a *= alpha;
		}

		return color;
	}

` + bumpedLightmapLighting + `

	vec4 LightmapPass(in vec4 color)
	{	
		if (renderLightmapsAsAlbedo == 1) {
			return color;
		}
		if (LightmapUV.x == -1) {
			return color;
		}

		vec3 normalTexel = vec3(0.5, 0.5, 1.0);
		if (bumpmap != 0) {
			normalTexel = texture(bumpmapSampler, UV).rgb;
		}

		return color * vec4(LightmapLighting(normalTexel), 1.0);
	}

    void main()
	{
		if (alphaTest == 1 && texture(albedoSampler, UV).a < alphaTestReference) {
			discard;
		}

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
	// Lightmaps are stored at half brightness, so that light can be up to twice as bright as a surface's texture
	const float lightmapScale = 2.0;

	// Flag that this material is in some way translucent
	uniform int hasTranslucentProperty;

	// Translucent variations
	uniform float alpha;
	uniform int translucent;

	// $alphatest discards the parts of the base texture whose alpha is below the reference
	uniform int alphaTest;
	uniform float alphaTestReference;

	// Debug Options
	uniform int renderLightmapsAsAlbedo;

	in vec2 UV;
	in vec2 LightmapUV;
	in float fadeAlpha;  // From vertex shader

    out vec4 frag_colour;

	vec4 AlbedoPass()
	{
		if (renderLightmapsAsAlbedo == 1) {
			return vec4(texture(lightmapSampler, LightmapUV).rgb * lightmapScale, 1.0);
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

		// Only $translucent uses the texture's alpha channel
		if (translucent == 0) {
			color.a = 1;
		}

		// $alpha fades the whole material. 0 means it is not set.
		if (alpha != 0) {
			color.a *= alpha;
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

		vec4 lightmapColor = vec4(texture(lightmapSampler, LightmapUV).rgb * lightmapScale, 1.0);

		return color * lightmapColor;
	}

    void main()
	{
		if (alphaTest == 1 && texture(albedoSampler, UV).a < alphaTestReference) {
			discard;
		}

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

// bumpedLightmapLighting is GLSL that lights a surface with its lightmaps, from the normal of its normal map if it has
// one. A shader that includes it declares lightmapSampler, lightmapScale and LightmapUV first.
//
//language=glsl
var bumpedLightmapLighting = `
	// The lightmap atlas has a page for the lightmap of each bump basis direction, below the surface's own lightmap
	uniform float lightmapPageHeight;
	// bumpmap is 0 to light a surface without its normal map, 1 to light it from the normal of its normal map, and 2
	// for a self-shadowing bump map ($ssbump), which holds how much each bump basis direction lights it
	uniform int bumpmap;

	// The directions each bump basis direction's lightmap is lit from, relative to the surface
	const vec3 bumpBasis[3] = vec3[3](
		vec3(0.81649661064147949, 0.0, 0.57735025882720947),
		vec3(-0.40824833512306213, 0.70710676908493042, 0.57735025882720947),
		vec3(-0.40824821591377258, -0.7071068286895752, 0.57735025882720947)
	);

	// LightmapLighting returns the light that reaches a surface, given its normal map's texel
	vec3 LightmapLighting(in vec3 normalTexel)
	{
		vec3 flatLight = texture(lightmapSampler, LightmapUV).rgb;
		if (bumpmap == 0) {
			return flatLight * lightmapScale;
		}

		vec3 light1 = texture(lightmapSampler, LightmapUV + vec2(0.0, lightmapPageHeight)).rgb;
		vec3 light2 = texture(lightmapSampler, LightmapUV + vec2(0.0, 2.0 * lightmapPageHeight)).rgb;
		vec3 light3 = texture(lightmapSampler, LightmapUV + vec2(0.0, 3.0 * lightmapPageHeight)).rgb;
		if (bumpmap == 2) {
			return (normalTexel.x * light1 + normalTexel.y * light2 + normalTexel.z * light3) * lightmapScale;
		}

		// Each direction lights the surface by how much its normal faces it
		vec3 normal = normalTexel * 2.0 - 1.0;
		vec3 weights = clamp(vec3(dot(normal, bumpBasis[0]), dot(normal, bumpBasis[1]), dot(normal, bumpBasis[2])), 0.0, 1.0);
		weights *= weights;
		float sum = weights.x + weights.y + weights.z;
		if (sum <= 0.0) {
			return flatLight * lightmapScale;
		}
		return (weights.x * light1 + weights.y * light2 + weights.z * light3) / sum * lightmapScale;
	}
`
