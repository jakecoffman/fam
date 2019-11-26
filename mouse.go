package fam

import "github.com/jakecoffman/cp"

type Mouse struct {
	worldPos cp.Vector
	body     *cp.Body
	joint    *cp.Constraint

	leftDownPos  *cp.Vector
	rightDownPos *cp.Vector
}

func (m *Mouse) New() {
	m.body = cp.NewKinematicBody()
}

func (m *Mouse) Update(dt float64) {
	newPoint := m.body.Position().Lerp(m.worldPos, 0.25)
	m.body.SetVelocityVector(newPoint.Sub(m.body.Position()).Mult(60.0))
	m.body.SetPosition(newPoint)
}