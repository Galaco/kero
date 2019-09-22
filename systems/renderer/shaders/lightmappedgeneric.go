package shaders

//language=glsl
var LightMappedGenericFragment = `
    #version 410

	uniform int useLightmap;

	uniform sampler2D albedoSampler;
	uniform sampler2D normalSampler;
	uniform sampler2D lightmapTextureSampler;

	in vec2 UV;
	in vec3 EyeDirection;
	in vec3 LightDirection;

    out vec4 frag_colour;

	vec4 GetAlbedo(in sampler2D sampler, in vec2 uv) 
	{
		return texture(sampler, uv).rgba;
	}

	float CalculateNormalFactor(in sampler2D sampler, vec2 uv)
	{
		vec3 L = normalize(LightDirection);
		vec3 N = normalize(texture(sampler, uv).xyz * 2.0 - 1.0);	// transform coordinates to -1-1 from 0-1

		return max(dot(N,L), 0.0);
	}

	vec4 GetSpecular(in sampler2D normalSampler, vec2 uv) {
		vec3 N = normalize(texture(normalSampler, uv).xyz * 2.0 - 1.0);	// transform coordinates to -1-1 from 0-1
		vec3 L = normalize(LightDirection);
		vec3 R = reflect(-L, N);
		vec3 V = normalize(EyeDirection);	

		// replace with texture where relevant
		vec3 specularSample = vec3(1.0);

		return max(pow(dot(R, V), 5.0), 0.0) * vec4(specularSample.xyz, 1.0);
	}

	// Lightmaps the face
	// Does nothing if lightmap was not defined
	vec4 GetLightmap(in sampler2D lightmap, in vec2 uv) 
	{
		return texture(lightmap, uv).rgba;
	}

    void main() 
	{
		float bumpFactor = CalculateNormalFactor(normalSampler, UV);
		vec4 diffuse = GetAlbedo(albedoSampler, UV);

		vec4 specular = GetSpecular(normalSampler, UV);

		frag_colour = diffuse + specular;
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
	layout(location = 4) in vec2 lightmapUV;

	out vec2 UV;
	out vec3 EyeDirection;
	out vec3 LightDirection;
	
	// temporary
	uniform vec3 lightPos = vec3(0.0, 0.0, 100.0);

	void calculateEyePosition() {
		// View space vertex position
		vec4 P = view * model * vec4(vertexPosition, 1.0);
		// Normal vector
		vec3 N = normalize(mat3(view * model) * vertexNormal);
		// Tangent vector
		vec3 T = normalize(mat3(view * model) * vertexTangent.xyz);
		// Bitangent vector
		vec3 B = cross(N, T);
		// Vector from target to viewer
		vec3 V = -P.xyz;

		EyeDirection = normalize(vec3(dot(V, T), dot(V, B), dot(V, N)));

		vec3 L = lightPos - P.xyz;
		LightDirection = normalize(vec3(dot(L, T), dot(L, B), dot(L, N)));
	}

    void main() {
		gl_Position = projection * view * model * vec4(vertexPosition, 1.0);

    	UV = vertexUV;

		// bump + specular related
		calculateEyePosition();
    }
` + "\x00"
