package fam

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
)

type Bomb struct {
	radius float64

	*Object
	Circle *cp.Circle

	state bombState
	time  float64
}

type bombState int

const (
	bombStateOk = iota
	bombStateBoom
	bombStateGone
)

const (
	bombTexture    = "bomb"
	bombPowTexture = "pow"
)

func NewBomb(pos cp.Vector, radius float64, space *cp.Space) *Bomb {
	p := &Bomb{
		Object: &Object{},
		state:  bombStateOk,
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
	p.Object.Update(g, dt)
	p.time += dt
	if p.time > 5 && p.state != bombStateBoom {
		p.state = bombStateBoom
		p.Circle.SetRadius(p.Circle.Radius() * explosionSizeIncrease)
	}
	if p.time > 5.2 && p.state == bombStateBoom {
		p.Circle.SetRadius(p.Circle.Radius() / explosionSizeIncrease)
	}
	if p.time > 6 {
		p.state = bombStateGone
		g.Space.RemoveShape(p.Shape)
		g.Space.RemoveBody(p.Body)
	}
}

func (p *Bomb) Draw(g *Game, alpha float64) {
	if p.state == bombStateGone {
		return
	}

	color := mgl32.Vec3{1, 1, 1}
	var texture *Texture2D

	switch p.state {
	case bombStateOk:
		texture = g.Texture(bombTexture)
		if int(p.time)%2 != 0 {
			// flash of grey representing bomb ticking ala Zelda bombs
			color = mgl32.Vec3{.5, .5, .5}
		}
	case bombStateBoom:
		texture = g.Texture(bombPowTexture)
	default:
		return
	}

	g.SpriteRenderer.DrawSprite(texture, p.SmoothPos(alpha), p.Size().Mul(2), p.Angle(), color)
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
		player.Circle.SetRadius(playerRadius)
		return true
	}

	return true
}
