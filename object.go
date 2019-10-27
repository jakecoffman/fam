package fam

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
)

var objectId int

func GetObjectId() int {
	objectId++
	return objectId
}

type Object struct {
	*cp.Body
	*cp.Shape

	lastPosition *cp.Vector
}

func (p *Object) Update(g *Game, dt float64) {
	pos := p.Position()
	p.lastPosition = &pos
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