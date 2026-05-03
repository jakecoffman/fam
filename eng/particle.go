package eng

import (
	"math/rand"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type Particle struct {
	Position, Velocity mgl32.Vec2
	Color              mgl32.Vec4
	Life               float64
}

func NewParticle() *Particle {
	return &Particle{
		Color: mgl32.Vec4{1, 1, 1, 1},
	}
}

type ParticleGenerator struct {
	particles []*Particle
	freeList  []int // indices of dead particles (stack)
	Amount    int
	Shader    *Shader
	Texture   *Texture2D
	VAO       uint32
}

func NewParticleGenerator(shader *Shader, texture *Texture2D, amount int) *ParticleGenerator {
	particleGenerator := &ParticleGenerator{
		Shader:  shader,
		Texture: texture,
		Amount:  amount,
	}

	var VBO uint32
	particleQuad := []float32{
		0, 1, 0, 1,
		1, 0, 1, 0,
		0, 0, 0, 0,

		0, 1, 0, 1,
		1, 1, 1, 1,
		1, 0, 1, 0,
	}
	gl.GenVertexArrays(1, &particleGenerator.VAO)
	gl.GenBuffers(1, &VBO)
	gl.BindVertexArray(particleGenerator.VAO)

	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(particleQuad)*4, gl.Ptr(particleQuad), gl.STATIC_DRAW)

	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 4, gl.FLOAT, false, 4*4, gl.PtrOffset(0))
	gl.BindVertexArray(0)

	particleGenerator.freeList = make([]int, 0, amount)
	for i := 0; i < amount; i++ {
		particleGenerator.particles = append(particleGenerator.particles, NewParticle())
		particleGenerator.freeList = append(particleGenerator.freeList, i)
	}

	return particleGenerator
}

func (p *ParticleGenerator) Update(dt float64, position, velocity mgl32.Vec2, newParticles int, offset mgl32.Vec2) {
	for i := 0; i < newParticles; i++ {
		unusedParticles := p.firstUnusedParticle()
		p.respawnParticle(p.particles[unusedParticles], position, velocity, offset)
	}
	for i := 0; i < p.Amount; i++ {
		particle := p.particles[i]
		if particle.Life <= 0 {
			continue
		}
		particle.Life -= dt
		if particle.Life <= 0 {
			// Particle just died — add to free list for reuse.
			p.freeList = append(p.freeList, i)
			continue
		}
		particle.Position = particle.Position.Add(particle.Velocity.Mul(float32(dt)))
		particle.Color = particle.Color.Sub(mgl32.Vec4{0, 0, 0, float32(dt) * 2.5})
	}
}

func (p *ParticleGenerator) Draw() {
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
	p.Shader.Use()
	// TODO: convert to instanced rendering for better GPU performance.
	for _, particle := range p.particles {
		if particle.Life > 0 {
			p.Shader.SetVec2f("offset", particle.Position)
			p.Shader.SetVec4f("color", particle.Color)
			p.Texture.Bind()
			gl.BindVertexArray(p.VAO)
			gl.DrawArrays(gl.TRIANGLES, 0, 6)
			gl.BindVertexArray(0)
		}
	}
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
}

// firstUnusedParticle returns the index of a dead particle using a free-list.
// Falls back to index 0 when the free list is empty (i.e. all particles are
// alive); it resets that particle so the oldest one is sacrificed for the new
// spawn. This is safe because particle 0 is arbitrary — it just prevents
// spawning from silently failing when the pool is exhausted.
func (p *ParticleGenerator) firstUnusedParticle() int {
	if len(p.freeList) > 0 {
		idx := p.freeList[len(p.freeList)-1]
		p.freeList = p.freeList[:len(p.freeList)-1]
		return idx
	}
	// Fallback: reuse the oldest particle (index 0).
	p.particles[0].Life = 0
	return 0
}

func (p *ParticleGenerator) respawnParticle(particle *Particle, position, velocity mgl32.Vec2, offset mgl32.Vec2) {
	random := (float32(rand.Intn(100)) - 50.0) / 10.0
	rColor := 0.5 + (float32(rand.Intn(100)) / 100.0)
	particle.Position = position.Add(mgl32.Vec2{offset.X() + random, offset.Y() + random})
	particle.Color = mgl32.Vec4{rColor, rColor, rColor, 1}
	particle.Life = 1
	particle.Velocity = velocity.Mul(0.1)
}
