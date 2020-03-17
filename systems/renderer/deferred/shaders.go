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
`

// language=glsl
var GeometryPassFragment = `
	#version 410
	
	uniform sampler2D albedoSampler;
	uniform int hasNormalSampler;
	uniform sampler2D normalSampler;
	
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

	vec3 GetNormal(in sampler2D sampler, in vec2 uv) 
	{
		if (hasNormalSampler > 0) {
			return texture(sampler, uv).rgb;
		}
		return normalize(Normal);
	}

	float GetSpecular(in sampler2D sampler, in vec2 uv) 
	{
		return 1;
		// return texture(sampler, uv).r;
	}
	
	void main() {
		PositionOut = Position;
		NormalOut = GetNormal(normalSampler, UV);
		AlbedoSpecularOut.rgb = GetAlbedo(albedoSampler, UV);
		AlbedoSpecularOut.a = GetSpecular(albedoSampler, UV);
	}
`

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
`

// language=glsl
var DirectionalLightPassFragment = `
	#version 410

	struct BaseLight
	{
		vec3 Color;
		float DiffuseIntensity;
	};

	struct DirectionalLight
	{
		BaseLight Base;
		vec3 AmbientColor;
		float AmbientIntensity;
		vec3 Direction;
	};
	
	in vec2 fsUv;

	uniform sampler2D uPositionTex;
	uniform sampler2D uNormalTex;
	uniform sampler2D uColorTex;

	uniform DirectionalLight directionalLight;

	out vec4 outColor;

	vec4 CalculateLightGeneric(BaseLight light, vec3 lightDirection) {
		vec3 worldPos = texture(uPositionTex, fsUv).xyz;
		vec3 normal = texture(uNormalTex, fsUv).xyz;

		vec4 ambientColor = vec4(directionalLight.AmbientColor * directionalLight.AmbientIntensity, 1.0f);
		float diffuseFactor = dot(normal, -lightDirection);
																								
		vec4 diffuseColor  = vec4(0, 0, 0, 0);
		vec4 specularColor = vec4(0, 0, 0, 0);

		if (diffuseFactor > 0) {
			diffuseColor = vec4(light.Color * light.DiffuseIntensity * diffuseFactor, 1.0f);

//			vec3 vertexToEye = normalize(gEyeWorldPos - worldPos);
//			vec3 lightReflect = normalize(reflect(lightDirection, normal));
//			float specularFactor = dot(vertexToEye, lightReflect);
//			if (specularFactor > 0) {
//				specularFactor = pow(specularFactor, gSpecularPower);
//				specularColor = vec4(Light.Color * texture(uColorTex, fsUV).a * specularFactor, 1.0f);
//			}
		}

		return (ambientColor + diffuseColor + specularColor);
	}

	vec4 CalculateDirectionalLight() {
		return CalculateLightGeneric(directionalLight.Base, directionalLight.Direction);
	}
	
	void main() {
		outColor = vec4(texture(uColorTex, fsUv).xyz, 1.0) * CalculateDirectionalLight();
	}
`

// language=glsl
var PointLightPassVertex = `
	#version 410

	layout(location = 0) in vec3 vsPos;
	out vec4 fsPos;

	uniform mat4 uVp;
	uniform float uLightRadius;
	uniform vec3 uLightPosition;

	void main()
	{
		vec4 pos = uVp * vec4((vsPos * uLightRadius) + uLightPosition, 1.0);

		gl_Position = pos;
		fsPos = pos;
	}
`

// language=glsl
var PointLightPassFragment = `
	#version 410

	uniform sampler2D uColorTex;
	uniform sampler2D uNormalTex;
	uniform sampler2D uPositionTex;

	out vec4 outColor;

	in vec4 fsPos;

	uniform float uLightRadius;
	uniform vec3 uLightPosition;
	uniform vec3 uLightColor;

	uniform vec3 uCameraPos;


	void main() {
		// get screen-space position of light sphere
		 // (remember to do perspective division.)
		vec2 uv = (fsPos.xy / fsPos.w) * 0.5 + 0.5;

		// now we can sample from the gbuffer for every fragment the light sphere covers.
		vec3 albedo = texture(uColorTex, uv).xyz;
		vec3 n = normalize(texture(uNormalTex, uv).xyz);
		vec3 pos = texture(uPositionTex, uv).xyz;

		vec3 lightToPosVector = pos.xyz - uLightPosition;
		float lightDist = length(lightToPosVector);  // position from light.
		vec3 l = -lightToPosVector / (lightDist);

		// implement fake z-test. If too far from light center, then 0.
		float ztest = step(0.0, uLightRadius - lightDist);

        // light attenuation.
		float d = lightDist / uLightRadius;
		float attenuation = 1.0 - d;
		vec3 v = normalize(uCameraPos - pos);
		vec3 h = normalize(l + v);

		vec3 color =
		// diffuse
		uLightColor * albedo.xyz * max(0.0, dot(n.xyz, l)) +
		// specular
		uLightColor * 0.4 * pow(max(0.0, dot(h, n)), 12.0);

		// finally ztest and attenuation.
		color *= ztest * attenuation;

		outColor = vec4(color, 1.0); // done!
	}
`