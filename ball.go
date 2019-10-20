package fam

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
)

type Ball struct {
	Texture *Texture2D
	Color   mgl32.Vec3

	*cp.Body
	*cp.Shape
	*cp.Circle
}

func NewBall(pos cp.Vector, radius float64, sprite *Texture2D) *Ball {
	ball := &Ball{
		Texture: sprite,
		Color:   mgl32.Vec3{1, 1, 1},
	}
	ball.Body = cp.NewBody(1, cp.MomentForCircle(1, radius, radius, cp.Vector{0, 0}))
	ball.Shape = cp.NewCircle(ball.Body, radius, cp.Vector{0, 0})
	ball.Shape.SetElasticity(0)
	ball.Shape.SetFriction(1)
	ball.Circle = ball.Shape.Class.(*cp.Circle)
	ball.Body.SetPosition(pos)
	return ball
}

func (b *Ball) Draw(renderer *SpriteRenderer, last *cp.Vector, alpha float64) {
	pos := b.Position()
	if last != nil {
		pos = pos.Mult(alpha).Add(last.Mult(1.0 - alpha))
	}
	bb := b.Shape.BB()
	size := mgl32.Vec2{
		float32(bb.R - bb.L),
		float32(bb.T - bb.B),
	}
	renderer.DrawSprite(b.Texture, V(pos), size, 0, b.Color)
}
