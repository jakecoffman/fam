package eng

import (
	"bytes"
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"sort"

	"github.com/go-gl/gl/v3.3-core/gl"
)

type ResourceManager struct {
	shaders    map[string]*Shader
	textures   map[string]*Texture2D
	textureIDs []uint32
}

func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		shaders:  map[string]*Shader{},
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
	img, err := loadImage(file)
	if err != nil {
		panic(err)
	}
	texture := NewTexture()
	texture.generateImage(img)
	r.textureIDs = append(r.textureIDs, texture.ID)
	r.textures[name] = texture
	return texture
}

func loadImage(path string) (image.Image, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decoding texture %q: %w", path, err)
	}
	return img, nil
}

func (r *ResourceManager) LoadTextureAtlas(files map[string]string) {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	images := make([]image.Image, len(names))
	for i, name := range names {
		img, err := loadImage(files[name])
		if err != nil {
			panic(err)
		}
		images[i] = img
	}
	var maxSize int32
	gl.GetIntegerv(gl.MAX_TEXTURE_SIZE, &maxSize)
	atlas, regions, err := packAtlas(images, int(maxSize), true)
	if err != nil {
		panic(err)
	}
	texture := NewTexture()
	texture.WrapS, texture.WrapT = gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE
	texture.generateImage(atlas)
	r.textureIDs = append(r.textureIDs, texture.ID)
	for i, name := range names {
		view := *texture
		view.Width, view.Height = regions[i].Dx(), regions[i].Dy()
		view.uvRect = atlasUV(regions[i], texture.Width, texture.Height)
		r.textures[name] = &view
	}
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
	for _, id := range r.textureIDs {
		gl.DeleteTextures(1, &id)
	}
	clear(r.shaders)
	clear(r.textures)
	r.textureIDs = nil
}
