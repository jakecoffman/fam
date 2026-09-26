package eng

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type SpriteRenderer struct {
	shader      *Shader
	quadVAO     uint32
	quadVBO     uint32
	instanceVBO uint32
	instances   []spriteInstance
	batches     []spriteBatch
}

type spriteInstance struct {
	model  mgl32.Mat4
	color  mgl32.Vec3
	uvRect mgl32.Vec4
}

type spriteBatch struct {
	texture *Texture2D
	start   int
	count   int
}

const spriteInstanceSize = (16 + 3 + 4) * 4

func NewSpriteRenderer(shader *Shader) *SpriteRenderer {
	renderer := &SpriteRenderer{shader: shader}
	renderer.initRenderData()
	return renderer
}

var (
	DefaultSpriteSize = mgl32.Vec2{10, 10}
	DefaultRotate     = 0.0
	DefaultColor      = mgl32.Vec3{1, 1, 1}
)

// DrawSprite queues a sprite, preserving draw order. Flush before using another renderer.
func (s *SpriteRenderer) DrawSprite(texture *Texture2D, position, size mgl32.Vec2, rotate float64, color mgl32.Vec3) {
	var model mgl32.Mat4
	model = mgl32.Translate3D(position.X(), position.Y(), 0)

	model = model.Mul4(mgl32.HomogRotate3D(float32(rotate), mgl32.Vec3{0, 0, 1}))
	model = model.Mul4(mgl32.Translate3D(-0.5*size.X(), -0.5*size.Y(), 0))
	model = model.Mul4(mgl32.Scale3D(size.X(), size.Y(), 1))

	if len(s.batches) == 0 || s.batches[len(s.batches)-1].texture.ID != texture.ID {
		s.batches = append(s.batches, spriteBatch{texture: texture, start: len(s.instances)})
	}
	s.instances = append(s.instances, spriteInstance{model: model, color: color, uvRect: texture.uvRect})
	s.batches[len(s.batches)-1].count++
}

func (s *SpriteRenderer) Flush() {
	if len(s.instances) == 0 {
		return
	}
	s.shader.Use()
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindVertexArray(s.quadVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, s.instanceVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(s.instances)*spriteInstanceSize, gl.Ptr(s.instances), gl.STREAM_DRAW)
	for _, batch := range s.batches {
		offset := batch.start * spriteInstanceSize
		for column := 0; column < 4; column++ {
			gl.VertexAttribPointer(uint32(1+column), 4, gl.FLOAT, false, spriteInstanceSize, gl.PtrOffset(offset+column*16))
		}
		gl.VertexAttribPointer(5, 3, gl.FLOAT, false, spriteInstanceSize, gl.PtrOffset(offset+64))
		gl.VertexAttribPointer(6, 4, gl.FLOAT, false, spriteInstanceSize, gl.PtrOffset(offset+76))
		batch.texture.Bind()
		gl.DrawArraysInstanced(gl.TRIANGLES, 0, 6, int32(batch.count))
	}
	gl.BindVertexArray(0)
	clear(s.batches)
	s.batches = s.batches[:0]
	s.instances = s.instances[:0]
}

func (s *SpriteRenderer) initRenderData() {
	vertices := []float32{
		0, 1, 0, 1,
		1, 0, 1, 0,
		0, 0, 0, 0,

		0, 1, 0, 1,
		1, 1, 1, 1,
		1, 0, 1, 0,
	}

	gl.GenVertexArrays(1, &s.quadVAO)
	gl.GenBuffers(1, &s.quadVBO)

	gl.BindBuffer(gl.ARRAY_BUFFER, s.quadVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.BindVertexArray(s.quadVAO)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 4, gl.FLOAT, false, 4*4, gl.PtrOffset(0))
	gl.GenBuffers(1, &s.instanceVBO)
	for attribute := uint32(1); attribute <= 6; attribute++ {
		gl.EnableVertexAttribArray(attribute)
		gl.VertexAttribDivisor(attribute, 1)
	}
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
}

func (s *SpriteRenderer) Destroy() {
	gl.DeleteVertexArrays(1, &s.quadVAO)
	gl.DeleteBuffers(1, &s.quadVBO)
	gl.DeleteBuffers(1, &s.instanceVBO)
}
