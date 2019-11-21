package eng

import (
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
)

type EntityID int

var nextEntityID EntityID

func NextEntityID() EntityID {
	nextEntityID++
	return nextEntityID
}

type Object struct {
	ID  EntityID
	Ptr interface{}
	*cp.Body
	*cp.Shape

	lastPosition *cp.Vector
}

type ObjectSystem struct {
	objects map[EntityID]*Object
	space *cp.Space
}

func NewObjectSystem(space *cp.Space) *ObjectSystem {
	return &ObjectSystem{
		space: space,
		objects: map[EntityID]*Object{},
	}
}

func (s *ObjectSystem) Add(ptr interface{}) *Object {
	p := &Object{
		ID:  NextEntityID(),
		Ptr: ptr,
	}
	s.objects[p.ID] = p
	return p
}

func (s *ObjectSystem) Get(id EntityID) *Object {
	return s.objects[id]
}

func (s *ObjectSystem) Remove(id EntityID) {
	log.Println("Removing game object", id)
	object := s.objects[id]
	delete(s.objects, id)
	object.Shape.UserData = nil
	object.Body.UserData = nil
	s.space.RemoveShape(object.Shape)
	s.space.RemoveBody(object.Body)
	object.Body.RemoveShape(object.Shape)
	object.Shape = nil
	object.Body = nil
}

func (s *ObjectSystem) Reset(space *cp.Space) {
	for id, o := range s.objects {
		s.space.RemoveShape(o.Shape)
		s.space.RemoveBody(o.Body)
		o.lastPosition = nil
		o.Ptr = nil
		o.Body = nil
		o.Shape = nil
		delete(s.objects, id)
	}
	s.space = space
}

func (s *ObjectSystem) Update(dt, worldWidth, worldHeight float64) {
	for _, p := range s.objects {
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
			// prevent smoothing
			p.lastPosition = nil
		}
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
