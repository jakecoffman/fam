package eng

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
)

type EntityId int

var objectId EntityId

func GetObjectId() EntityId {
	objectId++
	return objectId
}

type Object struct {
	ID EntityId
	Class interface{}
	*cp.Body
	*cp.Shape

	lastPosition *cp.Vector
}

func NewObject(class interface{}) *Object {
	return &Object{
		ID: GetObjectId(),
		Class: class,
	}
}

func (p *Object) Update(dt, worldWidth, worldHeight float64) {
	pos := p.Position()
	p.lastPosition = &pos

	bb := p.BB()
	if bb.R < 0 {
		pos.X = worldWidth+(bb.R-bb.L)/2
	}
	if bb.L > worldWidth {
		pos.X = -(bb.R-bb.L)/2
	}
	if bb.T < 0 {
		pos.Y = worldHeight-(bb.B-bb.T)/2
	}
	if bb.B > worldHeight {
		pos.Y = (bb.B-bb.T)/2
	}
	if !pos.Equal(p.Position()) {
		p.SetPosition(pos)
		// prevent smoothing
		p.lastPosition = nil
	}
}

func (p *Object) SmoothPos(alpha float64) mgl32.Vec2 {
	pos := p.Position()
	if p.lastPosition != nil {
		pos = pos.Mult(alpha).Add(p.lastPosition.Mult(1.0 - alpha))
	}
	return V(pos)
}

func (p *Object) Size() mgl32.Vec2 {
	bb := p.Shape.BB()
	return mgl32.Vec2{
		float32(bb.R-bb.L),
		float32(bb.T-bb.B),
	}
}
