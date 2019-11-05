package fam

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/fam/eng"
	"log"
)

const (
	MaxBananas = 100
)

type BananaSystem struct {
	active  int
	texture *eng.Texture2D
	game    *Game
	bananas [MaxBananas]Banana
	renderer *eng.SpriteRenderer
}

type Banana struct {
	*eng.Object
}

func (p *Banana) Update(dt float64) {
	p.Object.Update(dt, worldWidth, worldHeight)
}

func NewBananaSystem(g *Game) *BananaSystem {
	return &BananaSystem{
		game:    g,
		texture: g.Texture("banana"),
		bananas: [100]Banana{},
		renderer: g.SpriteRenderer,
	}
}

func (s *BananaSystem) Update(dt float64) {
	for i := 0; i < s.active; i++ {
		s.bananas[i].Update(dt)
	}
}

func (s *BananaSystem) Draw(alpha float64) {
	for i := 0; i < s.active; i++ {
		p := s.bananas[i]
		s.renderer.DrawSprite(s.texture, p.SmoothPos(alpha), p.Size(), p.Angle(), mgl32.Vec3{1, 1, 1})
	}
}

func (s *BananaSystem) Add() *eng.Object {
	if s.active >= MaxBananas {
		return nil
	}
	p := &s.bananas[s.active]
	s.active++
	p.Object = eng.NewObject(p)

	const (
		bananaMass = 10
		bananaRadius = 20
	)
	p.Body = cp.NewBody(bananaMass, cp.MomentForCircle(bananaMass, bananaRadius, bananaRadius, cp.Vector{0, 0}))
	p.Shape = cp.NewCircle(p.Body, bananaRadius, cp.Vector{0, 0})
	p.Shape.SetElasticity(0)
	p.Shape.SetFriction(10)

	// for consummation
	p.Shape.SetCollisionType(collisionBanana)
	p.Shape.SetFilter(PlayerFilter)

	p.Shape.UserData = p.ID
	s.game.Space.AddBody(p.Body)
	s.game.Space.AddShape(p.Shape)
	return p.Object
}

func (s *BananaSystem) Get(id eng.EntityId) (*eng.Object, int) {
	for i := 0; i < s.active; i++ {
		if s.bananas[i].ID == id {
			return s.bananas[i].Object, i
		}
	}
	return nil, -1
}

func (s *BananaSystem) Remove(index int) {
	if index >= s.active {
		log.Panic("Removing banana already removed")
	}
	s.active--
	s.bananas[s.active], s.bananas[index] = s.bananas[index], s.bananas[s.active]
	banana := &s.bananas[s.active]
	banana.Shape.UserData = nil
	s.game.Space.RemoveShape(banana.Shape)
	s.game.Space.RemoveBody(banana.Body)
	banana.Body.RemoveShape(banana.Shape)
	banana.Shape = nil
	banana.Body = nil
}

func swap(a, b *Banana) {
	*a, *b = *b, *a
}

func (s *BananaSystem) Reset() {
	for i := 0; i < s.active; i++ {
		s.game.Space.RemoveShape(s.bananas[i].Shape)
		s.game.Space.RemoveBody(s.bananas[i].Body)
	}
	s.active = 0
}

func BananaPreSolve(arb *cp.Arbiter, space *cp.Space, data interface{}) bool {
	game := data.(*Game)

	a, b := arb.Shapes()
	bid := a.UserData.(eng.EntityId)

	switch b.UserData.(type) {
	case *Player:
		player := b.UserData.(*Player)
		player.Circle.SetRadius(player.Circle.Radius() * 1.1)

		space.AddPostStepCallback(func(s *cp.Space, a interface{}, b interface{}) {
			_, index := game.Bananas.Get(bid)
			game.Bananas.Remove(index)
		}, nil, nil)

		return false
	}

	return true
}
