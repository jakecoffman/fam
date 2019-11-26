package fam

import (
	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/eng"
)

// Walls have no body of their own, they use the global static body, so this does not use eng.Object

const maxWalls = 1000

type WallSystem struct {
	game   *Game
	walls  map[eng.EntityID]*Wall
	active int
	pool   [maxWalls]Wall
}

type Wall struct {
	*cp.Segment
	ID eng.EntityID
}

func NewWallSystem(g *Game) *WallSystem {
	s := &WallSystem{
		game:  g,
		walls: map[eng.EntityID]*Wall{},
		pool: [maxWalls]Wall{},
	}
	wallPreSolve := func(arb *cp.Arbiter, space *cp.Space, data interface{}) bool {
		// allow jumping up through platforms
		if arb.Normal().Dot(cp.Vector{0, -1}) < 0 {
			return arb.Ignore()
		}
		return true
	}
	g.Space.NewWildcardCollisionHandler(collisionWall).PreSolveFunc = wallPreSolve
	return s
}

const (
	wallWidth    = 10
	wallFriction = 100
)

func (s *WallSystem) Add(a, b cp.Vector) *Wall {
	if s.active >= maxWalls {
		return &s.pool[s.active-1]
	}
	seg := cp.NewSegment(s.game.Space.StaticBody, a, b, wallWidth)
	seg.SetElasticity(1)
	seg.SetFriction(wallFriction)
	seg.SetCollisionType(collisionWall)
	// don't add to space because we might be in a callback
	p := &s.pool[s.active]
	p.ID = eng.NextEntityID()
	p.Segment = seg.Class.(*cp.Segment)
	seg.UserData = p.ID
	s.walls[p.ID] = p
	s.active++
	return p
}

func (s *WallSystem) Remove(id eng.EntityID) {
	// since the body is the static body we only remove the shape from the space
	s.game.Space.RemoveShape(s.walls[id].Shape)
	delete(s.walls, id)
	s.active--
	for i, p := range s.pool {
		if s.pool[i].ID == p.ID {
			s.pool[s.active], s.pool[i] = s.pool[i], s.pool[s.active]
			break
		}
	}
}

func (s *WallSystem) Draw(alpha float64) {
	for i := 0; i < s.active; i++ {
		w := s.pool[i]
		s.game.CPRenderer.DrawFatSegment(w.A(), w.B(), w.Radius(), eng.DefaultOutline, eng.DefaultFill)
	}
}
