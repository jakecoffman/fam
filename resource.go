package fam

import (
	"io/ioutil"
	"os"

	"github.com/go-gl/gl/v3.3-core/gl"
)

// Singleton
type ResourceManager struct {
	shaders map[string]*Shader
	textures map[string]*Texture2D
}

func NewResourceManager() *ResourceManager {
	return  &ResourceManager{
		shaders: map[string]*Shader{},
		textures: map[string]*Texture2D{},
	}
}

func (r *ResourceManager) LoadShader(vertexPath, fragmentPath, name string) *Shader {
	var vertexCode, fragmentCode string

	{
		bytes, err := ioutil.ReadFile(vertexPath)
		if err != nil {
			panic(err)
		}
		vertexCode = string(bytes)

		bytes, err = ioutil.ReadFile(fragmentPath)
		if err != nil {
			panic(err)
		}
		fragmentCode = string(bytes)
	}

	shader := NewShader(vertexCode, fragmentCode)
	r.shaders[name] = shader
	return shader
}

func (r *ResourceManager) Shader(name string) *Shader {
	shader, ok := r.shaders[name]
	if !ok {
		panic("Shader not found")
	}
	return shader
}

func (r *ResourceManager) LoadTexture(file string, name string) *Texture2D {
	texture := NewTexture()
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	texture.Generate(f)
	r.textures[name] = texture
	return texture
}

func (r *ResourceManager) Texture(name string) *Texture2D {
	t, ok := r.textures[name]
	if !ok {
		panic("Texture '" + name + "' not found")
	}
	return t
}

func (r *ResourceManager) Clear() {
	for _, shader := range r.shaders {
		gl.DeleteProgram(shader.ID)
	}
	for _, texture := range r.textures {
		gl.DeleteTextures(1, &texture.ID)
	}
}
