package fam

import (
	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/eng"
)

// Walls have no body of their own, they use the global static body, so this does not use Object
type WallSystem struct {
	*System

	game             *Game
	drawingWallShape *Wall
}

type Wall struct {
	*cp.Segment
	ID eng.EntityID
}

func NewWallSystem(g *Game) *WallSystem {
	const maxWalls = 1000
	s := &WallSystem{
		System: NewSystem(Wall{}, maxWalls),
		game:   g,
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
	// don't add to space because we might be in a callback
	ptr, ok := s.System.Add()
	p := ptr.(*Wall)
	if !ok {
		return p
	}

	seg := cp.NewSegment(s.game.Space.StaticBody, a, b, wallWidth)
	seg.SetElasticity(1)
	seg.SetFriction(wallFriction)
	seg.SetCollisionType(collisionWall)
	p.Segment = seg.Class.(*cp.Segment)
	seg.UserData = p.ID
	return p
}

func (s *WallSystem) Remove(id eng.EntityID) {
	// since the body is the static body we only remove the shape from the space
	s.game.Space.RemoveShape(s.Get(id).(*Wall).Shape)
	s.System.Remove(id)
}

func (s *WallSystem) Update(dt float64) {
	if s.game.mouse.leftDownPos != nil {
		s.drawingWallShape.SetEndpoints(*s.game.mouse.leftDownPos, s.game.mouse.worldPos)
	} else if s.drawingWallShape != nil {
		s.game.Space.AddShape(s.drawingWallShape.Shape)
		s.drawingWallShape = nil
	}
}

func (s *WallSystem) Draw(alpha float64) {
	pool := s.pool.([]Wall)
	for i := 0; i < s.active; i++ {
		w := &pool[i]
		s.game.CPRenderer.DrawFatSegment(w.A(), w.B(), w.Radius(), eng.DefaultOutline, eng.DefaultFill)
	}
}
