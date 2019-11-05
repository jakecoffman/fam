package fam

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/fam/eng"
	"log"
)

const MaxBombs = 100

type BombSystem struct {
	active   int
	game     *Game
	bombs    [MaxBombs]Bomb
	renderer *eng.SpriteRenderer

	texture, powTexture *eng.Texture2D
}

func NewBombSystem(g *Game) *BombSystem {
	const (
		bombTexture    = "bomb"
		bombPowTexture = "pow"
	)
	return &BombSystem{
		game:       g,
		texture:    g.Texture(bombTexture),
		powTexture: g.Texture(bombPowTexture),
		bombs:      [MaxBombs]Bomb{},
		renderer:   g.SpriteRenderer,
	}
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
	if s.active >= MaxBombs {
		return s.bombs[s.active-1].Object
	}
	p := &s.bombs[s.active]
	s.active++
	p.Object = eng.NewObject(p)

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

func (s *BombSystem) Get(id eng.EntityId) (ptr *eng.Object, index int) {
	for i := 0; i < s.active; i++ {
		if s.bombs[i].ID == id {
			return s.bombs[i].Object, i
		}
	}
	return nil, -1
}

func (s *BombSystem) Remove(index int) {
	if index >= s.active {
		log.Panic("Removing bomb already removed")
	}
	s.active--
	s.bombs[s.active], s.bombs[index] = s.bombs[index], s.bombs[s.active]
	bomb := &s.bombs[s.active]
	bomb.Shape.UserData = nil
	s.game.Space.RemoveShape(bomb.Shape)
	s.game.Space.RemoveBody(bomb.Body)
	bomb.Body.RemoveShape(bomb.Shape)
	bomb.Shape = nil
	bomb.Body = nil
}

func (s *BombSystem) Reset() {
	for i := 0; i < s.active; i++ {
		s.game.Space.RemoveShape(s.bombs[i].Shape)
		s.game.Space.RemoveBody(s.bombs[i].Body)
	}
	s.active = 0
}

func (s *BombSystem) Update(dt float64) {
	for i := 0; i < s.active; i++ {
		s.bombs[i].Update(dt)
		if s.bombs[i].state == bombStateGone {
			// avoid mutating array in for loop
			defer s.Remove(i)
		}
	}
}

func (s *BombSystem) Draw(alpha float64) {
	for i := 0; i < s.active; i++ {
		p := s.bombs[i]
		if p.state == bombStateGone {
			return
		}

		color := mgl32.Vec3{1, 1, 1}
		var texture *eng.Texture2D

		switch p.state {
		case bombStateOk:
			texture = s.texture
			if int(p.time)%2 != 0 {
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

const explosionSizeIncrease = 10

func (p *Bomb) Update(dt float64) {
	if p.state == bombStateGone {
		return
	}
	p.Object.Update(dt, worldWidth, worldHeight)
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
		player.Circle.SetRadius(playerRadius)
		return true
	}

	return true
}
