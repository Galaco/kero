package graphics

import (
	"github.com/galaco/gosigl"
)

type Shader struct {
	context  gosigl.Context
	uniforms map[string]int32
}

func (shader *Shader) Add(shaderType gosigl.ShaderType, code string) error {
	return shader.context.AddShader(code, shaderType)
}

func (shader *Shader) Finish() {
	shader.context.Finalize()
}

func (shader *Shader) Bind() {
	shader.context.UseProgram()
}

func (shader *Shader) GetUniform(name string) int32 {
	if _, ok := shader.uniforms[name]; !ok {
		shader.uniforms[name] = shader.context.GetUniform(name)
	}
	return shader.uniforms[name]
}

func NewShader() *Shader {
	return &Shader{
		context:  gosigl.NewShader(),
		uniforms: map[string]int32{},
	}
}
