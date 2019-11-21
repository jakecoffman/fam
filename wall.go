package fam

import (
	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/eng"
)

// Walls have no body of their own, they use the global static body, so this does not use eng.Object

type WallSystem struct {
	game  *Game
	walls map[eng.EntityID]*Wall
}

type Wall struct {
	*cp.Segment
	ID eng.EntityID
}

func NewWallSystem(g *Game) *WallSystem {
	return &WallSystem{
		game:  g,
		walls: map[eng.EntityID]*Wall{},
	}
}

const (
	wallWidth    = 10
	wallFriction = 100
)

func (s *WallSystem) Add(a, b cp.Vector) *Wall {
	seg := cp.NewSegment(s.game.Space.StaticBody, a, b, wallWidth)
	seg.SetElasticity(1)
	seg.SetFriction(wallFriction)
	seg.SetCollisionType(collisionWall)
	// don't add to space because we might be in a callback
	p := &Wall{
		ID:      eng.NextEntityID(),
		Segment: seg.Class.(*cp.Segment),
	}
	seg.UserData = p.ID
	s.walls[p.ID] = p
	return p
}

func (s *WallSystem) Remove(id eng.EntityID) {
	// since the body is the static body we only remove the shape from the space
	s.game.Space.RemoveShape(s.walls[id].Shape)
	delete(s.walls, id)
}

func (s *WallSystem) Reset() {
	s.walls = map[eng.EntityID]*Wall{}
	wallPreSolve := func(arb *cp.Arbiter, space *cp.Space, data interface{}) bool {
		// allow jumping up through platforms
		if arb.Normal().Dot(cp.Vector{0, -1}) < 0 {
			return arb.Ignore()
		}
		return true
	}
	s.game.Space.NewWildcardCollisionHandler(collisionWall).PreSolveFunc = wallPreSolve
}

func (s *WallSystem) Draw(alpha float64) {
	for _, w := range s.walls {
		s.game.CPRenderer.DrawFatSegment(w.A(), w.B(), w.Radius(), eng.DefaultOutline, eng.DefaultFill)
	}
}
