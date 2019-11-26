package fam

import (
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/eng"
)

type BombSystem struct {
	game     *Game
	bombs    map[eng.EntityID]*Bomb
	renderer *eng.SpriteRenderer

	texture, powTexture *eng.Texture2D
}

func NewBombSystem(g *Game) *BombSystem {
	const (
		bombTexture    = "bomb"
		bombPowTexture = "pow"
	)
	s := &BombSystem{
		game:       g,
		texture:    g.Texture(bombTexture),
		powTexture: g.Texture(bombPowTexture),
		bombs:      map[eng.EntityID]*Bomb{},
		renderer:   g.SpriteRenderer,
	}
	bombCollisionHandler := s.game.Space.NewCollisionHandler(collisionBomb, collisionPlayer)
	bombCollisionHandler.PreSolveFunc = func(arb *cp.Arbiter, space *cp.Space, data interface{}) bool {
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
	return s
}

type Bomb struct {
	*eng.Object
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

func (s *BombSystem) Add() *eng.Object {
	p := &Bomb{}
	p.Object = s.game.Objects.Add(s.game.Space)
	s.bombs[p.ID] = p

	p.state = bombStateOk
	p.time = 0
	const radius = 20
	p.Body = cp.NewBody(1, cp.MomentForCircle(1, radius, radius, cp.Vector{0, 0}))
	// the bomb body is smaller because of the wick, so make it a little smaller
	p.Shape = cp.NewCircle(p.Body, radius, cp.Vector{-radius, radius})
	p.Shape.SetElasticity(1)
	p.Shape.SetFriction(1)

	p.Shape.SetCollisionType(collisionBomb)
	p.Shape.SetFilter(PlayerFilter)
	p.Shape.UserData = p

	p.Circle = p.Shape.Class.(*cp.Circle)
	s.game.Space.AddBody(p.Body)
	s.game.Space.AddShape(p.Shape)
	return p.Object
}

func (s *BombSystem) Remove(id eng.EntityID) {
	log.Println("Removing bomb", id)
	s.game.Objects.Remove(id)
	delete(s.bombs, id)
}

func (s *BombSystem) Update(dt float64) {
	for _, p := range s.bombs {
		p.time += dt

		const explosionSizeIncrease = 10
		if p.time > 5 && p.state != bombStateBoom {
			p.state = bombStateBoom
			p.Circle.SetRadius(p.Circle.Radius() * explosionSizeIncrease)
		}
		if p.time > 5.2 && p.state == bombStateBoom {
			p.Circle.SetRadius(p.Circle.Radius() / explosionSizeIncrease)
		}
		if p.time > 6 {
			p.state = bombStateGone
			s.Remove(p.ID)
		}
	}
}

func (s *BombSystem) Draw(alpha float64) {
	for _, p := range s.bombs {
		if p.state == bombStateGone {
			return
		}

		color := mgl32.Vec3{1, 1, 1}
		var texture *eng.Texture2D

		switch p.state {
		case bombStateOk:
			texture = s.texture
			if int(p.time*3)%2 != 0 {
				// flash of grey representing bomb ticking ala Zelda bombs
				color = mgl32.Vec3{.5, .5, .5}
			}
		case bombStateBoom:
			texture = s.powTexture
		default:
			return
		}

		s.renderer.DrawSprite(texture, p.SmoothPos(alpha), p.Size().Mul(2), p.Angle(), color)
	}
}
