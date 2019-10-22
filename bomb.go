package fam

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
)

type Bomb struct {
	Texture, Boom *Texture2D
	radius float64

	*cp.Body
	*cp.Shape
	*cp.Circle

	state bombState
	time  float64
}

type bombState int
const (
	bombStateOk = iota
	bombStateBoom
	bombStateGone
)

func NewBomb(pos cp.Vector, radius float64, sprite, boom *Texture2D, space *cp.Space) *Bomb {
	p := &Bomb{
		Texture: sprite,
		Boom: boom,
		state: bombStateOk,
	}
	p.Body = cp.NewBody(1, cp.MomentForCircle(1, radius, radius, cp.Vector{0, 0}))
	// the bomb body is smaller because of the wick, so make it a little smaller
	p.Shape = cp.NewCircle(p.Body, radius, cp.Vector{-radius, radius})
	p.Shape.SetElasticity(1)
	p.Shape.SetFriction(0)

	p.Shape.SetCollisionType(collisionBomb)
	p.Shape.SetFilter(PlayerFilter)
	p.Shape.UserData = p

	p.Circle = p.Shape.Class.(*cp.Circle)
	p.Body.SetPosition(pos)
	space.AddBody(p.Body)
	space.AddShape(p.Shape)
	return p
}

const explosionSizeIncrease = 10

func (p *Bomb) Update(g *Game, dt float64) {
	if p.state == bombStateGone {
		return
	}
	p.time += dt
	if p.time > 5 && p.state != bombStateBoom {
		p.state = bombStateBoom
		p.Circle.SetRadius(p.Circle.Radius()*explosionSizeIncrease)
	}
	if p.time > 5.2 && p.state == bombStateBoom {
		p.Circle.SetRadius(p.Circle.Radius()/explosionSizeIncrease)
	}
	if p.time > 6 {
		p.state = bombStateGone
		g.Space.RemoveShape(p.Shape)
		g.Space.RemoveBody(p.Body)
	}
}

func (p *Bomb) Draw(renderer *SpriteRenderer, alpha float64) {
	if p.state == bombStateGone {
		return
	}

	pos := p.Position()
	//pos = pos.Mult(alpha).Add(p.LastPosition.Mult(1.0 - alpha))
	bb := p.Shape.BB()

	switch p.state {
	case bombStateOk:
		size := mgl32.Vec2{
			// doubling the size of bomb
			float32(bb.R - bb.L)*2,
			float32(bb.T - bb.B)*2,
		}
		if int(p.time) % 2 == 0 {
			renderer.DrawSprite(p.Texture, V(pos), size, p.Angle(), mgl32.Vec3{1, 1, 1})
		} else {
			renderer.DrawSprite(p.Texture, V(pos), size, p.Angle(), mgl32.Vec3{.5, .5, .5})
		}
	case bombStateBoom:
		size := mgl32.Vec2{
			// shrinking explosion sprite
			float32(bb.R - bb.L)*2,
			float32(bb.T - bb.B)*2,
		}
		renderer.DrawSprite(p.Boom, V(pos), size, p.Angle(), mgl32.Vec3{1, 1, 1})
	default:
		return
	}
}

func BombPreSolve(arb *cp.Arbiter, space *cp.Space, data interface{}) bool {
	a, b := arb.Shapes()

	bomb := a.UserData.(*Bomb)
	if bomb.state != bombStateBoom {
		return true
	}

	switch b.UserData.(type) {
	case *Player:
		player := b.UserData.(*Player)
		player.SetRadius(playerRadius)
		return true
	}

	return true
}
