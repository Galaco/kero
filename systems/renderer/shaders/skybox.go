package shaders

//language=glsl
var SkyboxFragment = `
    #version 410

	in vec3 UV;

    out vec4 frag_colour;

	uniform samplerCube albedoSampler;

    void main() {
		// Output color = color of the texture at the specified UV
		frag_colour = texture( albedoSampler, UV );
    }
` + "\x00"

// language=glsl
var SkyboxVertex = `
    #version 410

	uniform mat4 projection;
	uniform mat4 view;
	uniform mat4 model;

    layout(location = 0) in vec3 vertexPosition;
	layout(location = 1) in vec2 vertexUV;
	layout(location = 2) in vec2 vertexNormal;

	out vec3 UV;

    void main() {
		vec4 WVP_Pos = (projection * view * model) * vec4(vertexPosition, 1.0);
    	gl_Position = WVP_Pos.xyww;
    	UV = vertexPosition;
    }
` + "\x00"
