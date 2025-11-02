package shaders

//language=glsl
var DebugVertex = `
    #version 410

    uniform mat4 projection;
    uniform mat4 view;
    uniform mat4 model;

    layout(location = 0) in vec3 position;
    layout(location = 1) in vec3 color;

    out vec3 fragColor;

    void main()
    {
        gl_Position = projection * view * model * vec4(position, 1.0);
        fragColor = color;
    }
`

//language=glsl
var DebugFragment = `
    #version 410

    in vec3 fragColor;

    out vec4 frag_colour;

    void main()
    {
        frag_colour = vec4(fragColor, 1.0);
    }
`
