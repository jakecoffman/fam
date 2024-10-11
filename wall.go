package fam

import (
	"github.com/jakecoffman/cp/v2"
	"github.com/jakecoffman/fam/eng"
)

type Wall struct {
	*cp.Segment
}

const (
	wallWidth    = 10
	wallFriction = 100
)

func NewWall(g *Game, a, b cp.Vector) *Wall {
	seg := cp.NewSegment(g.Space.StaticBody, a, b, wallWidth)
	seg.SetElasticity(1)
	seg.SetFriction(wallFriction)
	seg.SetCollisionType(collisionWall)
	// don't add to space because we might be in a callback
	return &Wall{
		seg.Class.(*cp.Segment),
	}
}

func WallPreSolve(arb *cp.Arbiter, space *cp.Space, data interface{}) bool {
	// allow jumping up through platforms
	if arb.Normal().Dot(cp.Vector{0, -1}) < 0 {
		return arb.Ignore()
	}

	return true
}

func (w *Wall) Draw(g *Game, alpha float64) {
	g.CPRenderer.DrawFatSegment(w.A(), w.B(), w.Radius(), eng.DefaultOutline, eng.DefaultFill)
}
