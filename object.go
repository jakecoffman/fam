package fam

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/eng"
)

type Object struct {
	*cp.Body
	*cp.Shape

	ID eng.EntityID

	lastPosition *cp.Vector
}

func (p *Object) Update(dt float64) {
	pos := p.Position()
	p.lastPosition = &pos

	bb := p.BB()
	if bb.R < 0 {
		pos.X = worldWidth + (bb.R-bb.L)/2
	}
	if bb.L > worldWidth {
		pos.X = -(bb.R - bb.L) / 2
	}
	if bb.T < 0 {
		pos.Y = worldHeight - (bb.B-bb.T)/2
	}
	if bb.B > worldHeight {
		pos.Y = (bb.B - bb.T) / 2
	}
	if !pos.Equal(p.Position()) {
		p.SetPosition(pos)
		// prevent smoothing this frame
		p.lastPosition = nil
	}
}

func (p *Object) SmoothPos(alpha float64) mgl32.Vec2 {
	if p == nil {
		return mgl32.Vec2{}
	}
	pos := p.Position()
	if p.lastPosition != nil {
		pos = pos.Mult(alpha).Add(p.lastPosition.Mult(1.0 - alpha))
	}
	return eng.V(pos)
}

func (p *Object) Size() mgl32.Vec2 {
	bb := p.Shape.BB()
	return mgl32.Vec2{
		float32(bb.R - bb.L),
		float32(bb.T - bb.B),
	}
}

func (p *Object) Remove(space *cp.Space) {
	if p == nil || p.Shape == nil {
		return
	}
	p.Shape.UserData = nil
	p.Body.UserData = nil
	space.RemoveShape(p.Shape)
	space.RemoveBody(p.Body)
	p.Body.RemoveShape(p.Shape)
	p.Shape = nil
	p.Body = nil
}
