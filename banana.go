package fam

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/eng"
)

type BananaSystem struct {
	*System

	texture  *eng.Texture2D
	game     *Game
	renderer *eng.SpriteRenderer
}

type Banana struct {
	Object
}

func NewBananaSystem(g *Game) *BananaSystem {
	const maxBanana = 100
	s := &BananaSystem{
		System:   NewSystem(Banana{}, maxBanana),
		game:     g,
		texture:  g.Texture("banana"),
		renderer: g.SpriteRenderer,
	}
	bananaCollisionHandler := s.game.Space.NewCollisionHandler(collisionBanana, collisionPlayer)
	bananaCollisionHandler.UserData = s.game
	bananaCollisionHandler.PreSolveFunc = func(arb *cp.Arbiter, space *cp.Space, data interface{}) bool {
		game := data.(*Game)

		a, b := arb.Shapes()
		bid := a.UserData.(eng.EntityID)

		switch b.UserData.(type) {
		case *Player:
			player := b.UserData.(*Player)
			player.Circle.SetRadius(player.Circle.Radius() * 1.1)

			space.AddPostStepCallback(func(s *cp.Space, a interface{}, b interface{}) {
				game.Bananas.Remove(bid)
			}, nil, nil)

			return false
		}

		return true
	}
	return s
}

func (s *BananaSystem) Update(dt float64) {
	bananas := s.pool.([]Banana)
	for i := 0; i < s.active; i++ {
		bananas[i].Update(dt)
	}
}

func (s *BananaSystem) Draw(alpha float64) {
	bananas := s.pool.([]Banana)
	for i := 0; i < s.active; i++ {
		p := &bananas[i]
		s.renderer.DrawSprite(s.texture, p.SmoothPos(alpha), p.Size(), p.Angle(), mgl32.Vec3{1, 1, 1})
	}
}

func (s *BananaSystem) Add() *Object {
	ptr, ok := s.System.Add()
	p := ptr.(*Banana)
	if !ok {
		return &p.Object
	}

	const (
		bananaMass   = 10
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
	return &p.Object
}

func (s *BananaSystem) Remove(id eng.EntityID) {
	s.System.Get(id).(*Banana).Remove(s.game.Space)
	s.System.Remove(id)
}
