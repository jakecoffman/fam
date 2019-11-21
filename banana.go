package fam

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/eng"
)

type BananaSystem struct {
	texture  *eng.Texture2D
	game     *Game
	bananas  map[eng.EntityID]*Banana
	renderer *eng.SpriteRenderer
}

type Banana struct {
	*eng.Object
}

func NewBananaSystem(g *Game) *BananaSystem {
	return &BananaSystem{
		game:     g,
		texture:  g.Texture("banana"),
		bananas:  map[eng.EntityID]*Banana{},
		renderer: g.SpriteRenderer,
	}
}

func (s *BananaSystem) Draw(alpha float64) {
	for _, p := range s.bananas {
		s.renderer.DrawSprite(s.texture, p.SmoothPos(alpha), p.Size(), p.Angle(), mgl32.Vec3{1, 1, 1})
	}
}

func (s *BananaSystem) Add() *eng.Object {
	p := &Banana{}
	p.Object = s.game.Objects.Add(p)
	s.bananas[p.ID] = p

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
	return p.Object
}

func (s *BananaSystem) Remove(id eng.EntityID) {
	s.game.Objects.Remove(id)
	delete(s.bananas, id)
}

func (s *BananaSystem) Reset() {
	for id := range s.bananas {
		delete(s.bananas, id)
	}
	bananaCollisionHandler := s.game.Space.NewCollisionHandler(collisionBanana, collisionPlayer)
	bananaCollisionHandler.UserData = s.game
	bananaPreSolve := func(arb *cp.Arbiter, space *cp.Space, data interface{}) bool {
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
	bananaCollisionHandler.PreSolveFunc = bananaPreSolve
}
