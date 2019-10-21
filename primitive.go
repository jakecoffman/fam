package fam

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type PrimitiveRenderer struct {
	shader   *Shader
	VAO, VBO uint32
}

func NewPrimitiveRenderer(shader *Shader) *PrimitiveRenderer {
	renderer := &PrimitiveRenderer{shader: shader}
	renderer.initRenderData()
	return renderer
}

func (s *PrimitiveRenderer) DrawPrimitive(position, size mgl32.Vec2, rotate float64, color mgl32.Vec3) {
	s.shader.Use()
	var model mgl32.Mat4
	model = mgl32.Translate3D(position.X(), position.Y(), 0)

	model = model.Mul4(mgl32.Translate3D(0.5*size.X(), 0.5*size.Y(), 0))
	model = model.Mul4(mgl32.HomogRotate3D(float32(rotate), mgl32.Vec3{0, 0, 1}))
	model = model.Mul4(mgl32.Translate3D(-0.5*size.X(), -0.5*size.Y(), 0))

	model = model.Mul4(mgl32.Scale3D(size.X(), size.Y(), 1))

	s.shader.SetMat4("model", model)
	s.shader.SetVec3f("primitiveColor", color)

	gl.BindVertexArray(s.VAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 3)
}

func (s *PrimitiveRenderer) initRenderData() {
	vertices := []float32{
		1, -1, 0, // top right
		-1, -1, 0, // top left
		0, 0, 0, // bottom
	}

	gl.GenVertexArrays(1, &s.VAO)
	gl.GenBuffers(1, &s.VBO)
	gl.BindVertexArray(s.VAO)

	gl.BindBuffer(gl.ARRAY_BUFFER, s.VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 4*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)
}

func (s *PrimitiveRenderer) Destroy() {
	gl.DeleteVertexArrays(1, &s.VAO)
}
