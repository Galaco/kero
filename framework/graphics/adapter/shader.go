package adapter

import "github.com/galaco/gosigl"

type Shader struct {
	context gosigl.Context
}

var currentShader *Shader

// CurrentShader returns the currently bound shader (if exists)
func CurrentShader() *Shader {
	return currentShader
}

func (shader *Shader) Add(shaderType gosigl.ShaderType, code string) error {
	return shader.context.AddShader(code, shaderType)
}

func (shader *Shader) Finish() {
	shader.context.Finalize()
}

func (shader *Shader) Bind() {
	shader.context.UseProgram()
	currentShader = shader
}

func (shader *Shader) GetUniform(name string) int32 {
	return shader.context.GetUniform(name)
}

func NewShader() *Shader {
	return &Shader{
		context: gosigl.NewShader(),
	}
}
