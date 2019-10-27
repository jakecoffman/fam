package eng

import (
	"log"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type Shader struct {
	ID uint32
}

func NewShader(vertexCode, fragmentCode string) *Shader {
	vertexShader := CompileShader(gl.VERTEX_SHADER, vertexCode)
	fragmentShader := CompileShader(gl.FRAGMENT_SHADER, fragmentCode)
	ID := LinkProgram(vertexShader, fragmentShader)
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return &Shader{
		ID: ID,
	}
}

func (s *Shader) Use() *Shader {
	gl.UseProgram(s.ID)
	return s
}

func (s *Shader) SetBool(name string, value bool) *Shader {
	if value {
		gl.Uniform1i(gl.GetUniformLocation(s.ID, gl.Str(name+"\x00")), 1)
	} else {
		gl.Uniform1i(gl.GetUniformLocation(s.ID, gl.Str(name+"\x00")), 0)
	}
	return s
}

func (s *Shader) SetInt(name string, value int) *Shader {
	gl.Uniform1i(gl.GetUniformLocation(s.ID, gl.Str(name+"\x00")), int32(value))
	return s
}

func (s *Shader) SetFloat(name string, value float64) *Shader {
	gl.Uniform1f(gl.GetUniformLocation(s.ID, gl.Str(name+"\x00")), float32(value))
	return s
}

func (s *Shader) SetVec2f(name string, value mgl32.Vec2) *Shader {
	gl.Uniform2f(gl.GetUniformLocation(s.ID, gl.Str(name+"\x00")), value.X(), value.Y())
	return s
}

func (s *Shader) SetVec3f(name string, value mgl32.Vec3) *Shader {
	gl.Uniform3f(gl.GetUniformLocation(s.ID, gl.Str(name+"\x00")), value.X(), value.Y(), value.Z())
	return s
}

func (s *Shader) SetVec4f(name string, value mgl32.Vec4) *Shader {
	gl.Uniform4f(gl.GetUniformLocation(s.ID, gl.Str(name+"\x00")), value.X(), value.Y(), value.Z(), value.W())
	return s
}

func (s *Shader) SetMat4(name string, value mgl32.Mat4) *Shader {
	gl.UniformMatrix4fv(gl.GetUniformLocation(s.ID, gl.Str(name+"\x00")), 1, false, &value[0])
	return s
}

func CheckGLErrors() {
	for err := gl.GetError(); err != 0; err = gl.GetError() {
		switch err {
		case gl.NO_ERROR:
			// ok
		case gl.INVALID_ENUM:
			panic("Invalid enum")
		case gl.INVALID_VALUE:
			panic("Invalid value")
		case gl.INVALID_OPERATION:
			panic("Invalid operation")
		case gl.INVALID_FRAMEBUFFER_OPERATION:
			panic("Invalid Framebuffer Operation")
		case gl.OUT_OF_MEMORY:
			panic("Out of memory")
		}
	}
}

func CheckError(obj uint32, status uint32, getiv func(uint32, uint32, *int32), getInfoLog func(uint32, int32, *int32, *uint8)) bool {
	var success int32
	getiv(obj, status, &success)

	if success == gl.FALSE {
		var length int32
		getiv(obj, gl.INFO_LOG_LENGTH, &length)

		info := strings.Repeat("\x00", int(length+1))
		getInfoLog(obj, length, nil, gl.Str(info))

		log.Println("GL Error:", info)
		return true
	}

	return false
}

func CompileShader(typ uint32, source string) uint32 {
	shader := gl.CreateShader(typ)

	sources, free := gl.Strs(source + "\x00")
	defer free()
	gl.ShaderSource(shader, 1, sources, nil)
	gl.CompileShader(shader)

	if CheckError(shader, gl.COMPILE_STATUS, gl.GetShaderiv, gl.GetShaderInfoLog) {
		panic("Error compiling shader")
	}

	return shader
}

func LinkProgram(vshader, fshader uint32) uint32 {
	p := gl.CreateProgram()

	gl.AttachShader(p, vshader)
	gl.AttachShader(p, fshader)

	gl.LinkProgram(p)

	if CheckError(p, gl.LINK_STATUS, gl.GetProgramiv, gl.GetProgramInfoLog) {
		panic("Error linking shader program")
	}

	return p
}

func SetAttribute(program uint32, name string, size int32, gltype uint32, stride int32, offset int) {
	var index = uint32(gl.GetAttribLocation(program, gl.Str(name+"\x00")))
	gl.EnableVertexAttribArray(index)
	gl.VertexAttribPointer(index, size, gltype, false, stride, gl.PtrOffset(offset))
	CheckGLErrors()
}
