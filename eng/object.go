package eng

import (
	"math"
	"sync/atomic"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp/v2"
)

var objectId int64

func GetObjectId() int64 {
	return atomic.AddInt64(&objectId, 1)
}

type Object struct {
	*cp.Body
	*cp.Shape

	// lastPosition/lastAngle capture the state at the start of each physics
	// sub-step (before cp.Space.Step). This is correct because Object.Update
	// is called once per sub-step, immediately before the step runs.
	lastPosition cp.Vector
	lastAngle    float64
	hasLast      bool
}

// Update captures last-frame state for interpolation and handles world wrapping.
// space is used to reindex the broadphase after a wrap teleport.
// Must be called once per physics sub-step, before Space.Step.
func (p *Object) Update(space *cp.Space, dt, worldWidth, worldHeight float64) {
	pos := p.Position()
	p.lastPosition = pos
	p.lastAngle = p.Body.Angle()
	p.hasLast = true

	bb := p.BB()
	if bb.R < 0 {
		pos.X += worldWidth
	}
	if bb.L > worldWidth {
		pos.X -= worldWidth
	}
	if bb.T < 0 {
		pos.Y += worldHeight
	}
	if bb.B > worldHeight {
		pos.Y -= worldHeight
	}
	if !pos.Equal(p.Position()) {
		p.SetPosition(pos)
		// After teleporting, snap the interpolation origin too so there's no
		// one-frame streak across the screen.
		p.lastPosition = pos
		space.ReindexShapesForBody(p.Body)
	}
}

func (p *Object) SmoothPos(alpha float64) mgl32.Vec2 {
	pos := p.Position()
	if p.hasLast {
		pos = pos.Mult(alpha).Add(p.lastPosition.Mult(1.0 - alpha))
	}
	return V(pos)
}

// SmoothAngle returns the interpolated angle between the last physics step and
// the current one, using shortest-arc lerp to avoid wrap-around artifacts.
func (p *Object) SmoothAngle(alpha float64) float64 {
	if !p.hasLast {
		return p.Body.Angle()
	}
	cur := p.Body.Angle()
	diff := cur - p.lastAngle
	// Wrap diff into [-π, π] for shortest-arc interpolation.
	for diff > math.Pi {
		diff -= 2 * math.Pi
	}
	for diff < -math.Pi {
		diff += 2 * math.Pi
	}
	return p.lastAngle + diff*alpha
}

func (p *Object) Size() mgl32.Vec2 {
	bb := p.Shape.BB()
	return mgl32.Vec2{
		float32(bb.R - bb.L),
		float32(bb.T - bb.B),
	}
}
