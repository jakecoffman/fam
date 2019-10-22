package fam

import (
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
	"math"
)

type Ball struct {
	Texture *Texture2D
	Color   mgl32.Vec3

	*cp.Body
	Control *cp.Body
	*cp.Shape
	*cp.Circle

	Joystick glfw.Joystick
}

func NewBall(pos cp.Vector, radius float64, sprite *Texture2D, space *cp.Space) *Ball {
	ball := &Ball{
		Texture: sprite,
		Color:   mgl32.Vec3{1, 1, 1},
	}
	ball.Body = cp.NewBody(1, cp.MomentForCircle(1, radius, radius, cp.Vector{0, 0}))
	ball.Shape = cp.NewCircle(ball.Body, radius, cp.Vector{0, 0})
	ball.Shape.SetElasticity(0)
	ball.Shape.SetFriction(1)
	ball.Circle = ball.Shape.Class.(*cp.Circle)
	ball.Body.SetPosition(pos)

	ball.Control = space.AddBody(cp.NewKinematicBody())
	pivot := space.AddConstraint(cp.NewPivotJoint2(ball.Control, ball.Body, cp.Vector{}, cp.Vector{}))
	pivot.SetMaxBias(0)
	pivot.SetMaxForce(10000)

	gear := space.AddConstraint(cp.NewGearJoint(ball.Control, ball.Body, 0.0, 1.0))
	gear.SetErrorBias(0) // attempt to fully correct the joint each step
	gear.SetMaxBias(1.2)
	gear.SetMaxForce(50000)

	space.AddBody(ball.Body)
	space.AddShape(ball.Shape)

	return ball
}

func (b *Ball) processInput(g *Game, dt float64) {
	//if g.state != stateActive {
	//	return
	//}

	velocity := playerVelocity * dt

	force := cp.Vector{}

	if b.Joystick > -1 {
		axes := glfw.GetJoystickAxes(b.Joystick)
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

	b.Control.SetVelocityVector(force)
}

func (b *Ball) Draw(renderer *SpriteRenderer, last *cp.Vector, alpha float64) {
	pos := b.Position()
	//if last != nil {
	//	pos = pos.Mult(alpha).Add(last.Mult(1.0 - alpha))
	//}
	bb := b.Shape.BB()
	size := mgl32.Vec2{
		float32(bb.R - bb.L),
		float32(bb.T - bb.B),
	}
	renderer.DrawSprite(b.Texture, V(pos), size, b.Angle(), b.Color)
}
