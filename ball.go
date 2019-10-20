package fam

import (
	"github.com/go-gl/mathgl/mgl32"
)

type Ball struct {
	*Object
	Radius float32
}

func NewBall(pos mgl32.Vec2, radius float32, velocity mgl32.Vec2, sprite *Texture2D) *Ball {
	ball := &Ball{}
	ball.Object = NewGameObject(pos, mgl32.Vec2{radius * 2, radius * 2}, sprite)
	ball.Color = mgl32.Vec3{1, 1, 1}
	ball.Velocity = velocity
	ball.Radius = radius
	return ball
}

func (b *Ball) Move(dt, windowWidth, windowHeight float32) mgl32.Vec2 {
	b.Position = b.Position.Add(b.Velocity.Mul(dt))
	if b.Position.X() <= 0 {
		b.Velocity = mgl32.Vec2{-b.Velocity.X(), b.Velocity.Y()}
		b.Position = mgl32.Vec2{0, b.Position.Y()}
	} else if b.Position.X() + b.Size.X() >= windowWidth {
		b.Velocity = mgl32.Vec2{-b.Velocity.X(), b.Velocity.Y()}
		b.Position = mgl32.Vec2{windowWidth-b.Size.X(), b.Position.Y()}
	}
	if b.Position.Y() <= 0 {
		b.Velocity = mgl32.Vec2{b.Velocity.X(), -b.Velocity.Y()}
		b.Position = mgl32.Vec2{b.Position.X(), 0}
	} else if b.Position.Y() + b.Size.Y() >= windowHeight {
		b.Velocity = mgl32.Vec2{b.Velocity.X(), -b.Velocity.Y()}
		b.Position = mgl32.Vec2{b.Position.X(), windowHeight-b.Size.Y()}
	}

	return b.Position
}

func (b *Ball) Reset(position, velocity mgl32.Vec2) {
	b.Position = position
	b.Velocity = velocity
}
