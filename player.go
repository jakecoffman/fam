package fam

import (
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
	"math"
)

type Player struct {
	Texture *Texture2D
	Color   mgl32.Vec3

	*cp.Body
	Control *cp.Body
	*cp.Shape
	*cp.Circle

	Joystick glfw.Joystick

	LastPosition cp.Vector
}

func NewPlayer(pos cp.Vector, radius float64, sprite *Texture2D, space *cp.Space) *Player {
	p := &Player{
		Texture: sprite,
		Color:   mgl32.Vec3{1, 1, 1},
	}
	p.Reset(pos, radius, space)

	return p
}

func (p *Player) Reset(pos cp.Vector, radius float64, space *cp.Space) {
	p.Body = cp.NewBody(1, cp.MomentForCircle(1, radius, radius, cp.Vector{0, 0}))
	p.Shape = cp.NewCircle(p.Body, radius, cp.Vector{0, 0})
	p.Shape.SetElasticity(0)
	p.Shape.SetFriction(1)

	p.Shape.SetFilter(cp.NewShapeFilter(uint(GetObjectId()), PlayerMaskBit, PlayerMaskBit))
	p.Shape.SetCollisionType(collisionPlayer)
	p.Shape.UserData = p

	p.Circle = p.Shape.Class.(*cp.Circle)
	p.Body.SetPosition(pos)

	p.Control = space.AddBody(cp.NewKinematicBody())
	pivot := space.AddConstraint(cp.NewPivotJoint2(p.Control, p.Body, cp.Vector{}, cp.Vector{}))
	pivot.SetMaxBias(0)
	pivot.SetMaxForce(10000)

	gear := space.AddConstraint(cp.NewGearJoint(p.Control, p.Body, 0.0, 1.0))
	gear.SetErrorBias(0) // attempt to fully correct the joint each step
	gear.SetMaxBias(1.2)
	gear.SetMaxForce(50000)

	space.AddBody(p.Body)
	space.AddShape(p.Shape)
}

func (p *Player) Update(g *Game, dt float64) {
	//if g.state != stateActive {
	//	return
	//}

	//p.LastPosition = p.Position()

	velocity := playerVelocity * dt

	force := cp.Vector{}

	if p.Joystick > -1 {
		axes := glfw.GetJoystickAxes(p.Joystick)
		if len(axes) == 0 {
			return
		}

		const sensitivity = 100
		force.X = math.Round(float64(axes[0])*sensitivity) / sensitivity * velocity
		force.Y = math.Round(float64(axes[1])*sensitivity) / sensitivity * -velocity
	} else {
		if g.Keys[glfw.KeyA] || g.Keys[glfw.KeyLeft] {
			force.X = -velocity
		}
		if g.Keys[glfw.KeyD] || g.Keys[glfw.KeyRight] {
			force.X = velocity
		}
		if g.Keys[glfw.KeyW] || g.Keys[glfw.KeyUp] {
			force.Y = -velocity
		}
		if g.Keys[glfw.KeyS] || g.Keys[glfw.KeyDown] {
			force.Y = velocity
		}
	}

	p.Control.SetVelocityVector(force)
}

func (p *Player) Draw(renderer *SpriteRenderer, alpha float64) {
	pos := p.Position()
	//pos = pos.Mult(alpha).Add(p.LastPosition.Mult(1.0 - alpha))
	bb := p.Shape.BB()
	size := mgl32.Vec2{
		float32(bb.R - bb.L),
		float32(bb.T - bb.B),
	}
	renderer.DrawSprite(p.Texture, V(pos), size, p.Angle(), p.Color)
}
