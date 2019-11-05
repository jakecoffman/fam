package eng

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
)

type System interface {
	Add() *Object
	Get(id EntityId) (ptr *Object, index int)
	Remove(index int)
	Reset()

	Update(dt float64)
	Draw(alpha float64)
}

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

	if pos.X < -5 {
		pos.X = worldWidth
	}
	if pos.X > worldWidth + 5 {
		pos.X = 0
	}
	if pos.Y < -5 {
		pos.Y = worldHeight
	}
	if pos.Y > worldHeight + 5 {
		pos.Y = 0
	}
	if !pos.Equal(p.Position()) {
		p.SetPosition(pos)
		// prevent smoothing
		p.lastPosition = &pos
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
