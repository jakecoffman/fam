package fam

import (
	"math"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/fam/eng"
)

type Player struct {
	Color   mgl32.Vec3

	*eng.Object
	Circle *cp.Circle

	Joystick glfw.Joystick

	lastPosition *cp.Vector

	remainingBoost float64
	grounded, lastJumpState bool
}

func NewPlayer(pos cp.Vector, radius float64, g *Game) *Player {
	p := &Player{
		Object: &eng.Object{},
		Color:   mgl32.Vec3{1, 1, 1},
	}
	p.Reset(pos, radius, g)

	return p
}

func (p *Player) Reset(pos cp.Vector, radius float64, g *Game) {
	p.Body = cp.NewBody(1, cp.MomentForCircle(1, radius, radius, cp.Vector{0, 0}))
	p.Body.SetVelocityUpdateFunc(playerUpdateVelocity(g, p))

	p.Shape = cp.NewCircle(p.Body, radius, cp.Vector{0, 0})
	p.Shape.SetElasticity(0)
	p.Shape.SetFriction(1)

	p.Shape.SetFilter(cp.NewShapeFilter(uint(eng.GetObjectId()), PlayerMaskBit, PlayerMaskBit))
	p.Shape.SetCollisionType(collisionPlayer)
	p.Shape.UserData = p

	p.Circle = p.Shape.Class.(*cp.Circle)
	p.Body.SetPosition(pos)

	g.Space.AddBody(p.Body)
	g.Space.AddShape(p.Shape)
}

func (p *Player) Update(g *Game, dt float64) {
	p.Object.Update(dt, worldWidth, worldHeight)

	var jumpState bool
	if p.Joystick > -1 {
		buttonBytes := glfw.GetJoystickButtons(p.Joystick)
		if len(buttonBytes) > 0 {
			jumpState = glfw.Action(buttonBytes[0]) == glfw.Press
		}
	} else {
		jumpState = g.Keys[glfw.KeySpace]
	}
	// If the jump key was just pressed this frame, jump!
	if jumpState && !p.lastJumpState && p.grounded {
		jumpV := -math.Sqrt(2.0 * JumpHeight * Gravity)
		p.SetVelocityVector(p.Velocity().Add(cp.Vector{0, jumpV}))

		p.remainingBoost = JumpBoostHeight / jumpV
	}
	p.remainingBoost -= dt
	p.lastJumpState = jumpState
}

func (p *Player) Draw(g *Game, alpha float64) {
	g.SpriteRenderer.DrawSprite(
		g.Texture("face"),
		p.SmoothPos(alpha),
		p.Size().Mul(1.1), // increase 10% to better fit hitbox
		p.Angle(),
		p.Color)
}

const (
	PlayerVelocity = 500.0

	PlayerGroundAccelTime = 0.1
	PlayerGroundAccel     = PlayerVelocity / PlayerGroundAccelTime

	PlayerAirAccelTime = 0.25
	PlayerAirAccel     = PlayerVelocity / PlayerAirAccelTime

	JumpHeight      = 250.0
	JumpBoostHeight = 955.0
	FallVelocity    = 900.0
	Gravity         = 2000.0

	joystickSensitivity = 100
)

func playerUpdateVelocity(g *Game, p *Player) func(*cp.Body, cp.Vector, float64, float64) {
	return func(body *cp.Body, gravity cp.Vector, damping, dt float64) {
		var jumpState bool
		var x float64

		if p.Joystick > -1 {
			axes := glfw.GetJoystickAxes(p.Joystick)
			if len(axes) < 2 {
				return
			}

			buttonBytes := glfw.GetJoystickButtons(p.Joystick)
			if len(buttonBytes) > 0 {
				jumpState = glfw.Action(buttonBytes[0]) == glfw.Press
			}
			x = math.Round(float64(axes[0])*joystickSensitivity) / joystickSensitivity
		} else {
			if g.Keys[glfw.KeyA] || g.Keys[glfw.KeyLeft] {
				x = -1
			}
			if g.Keys[glfw.KeyD] || g.Keys[glfw.KeyRight] {
				x = 1
			}
			jumpState = g.Keys[glfw.KeySpace]
		}

		// Grab the grounding normal from last frame
		groundNormal := cp.Vector{}
		body.EachArbiter(func(arb *cp.Arbiter) {
			n := arb.Normal()

			if n.Y > groundNormal.Y {
				groundNormal = n
			}
		})

		p.grounded = groundNormal.Y > 0
		if groundNormal.Y < 0 {
			p.remainingBoost = 0
		}

		// Do a normal-ish update
		boost := jumpState && p.remainingBoost > 0
		var grav cp.Vector
		if !boost {
			grav = gravity
		}
		body.UpdateVelocity(grav, damping, dt)

		// Target horizontal speed for air/ground control
		targetVx := PlayerVelocity * x

		// Update the surface velocity and friction
		// Note that the "feet" move in the opposite direction of the player.
		v := p.Velocity()

		//surfaceV := cp.Vector{-targetVx, v.Y}
		//p.Shape.SetSurfaceV(surfaceV)

		//if p.grounded {
		//	p.Shape.SetFriction(PlayerGroundAccel / Gravity)
		//} else {
		//	p.Shape.SetFriction(0)
		//}

		// Apply air control if not grounded

		if !p.grounded {
			p.SetVelocity(cp.LerpConst(v.X, targetVx, PlayerAirAccel*dt), v.Y)
		} else {
			p.SetVelocity(targetVx, v.Y)
		}

		v = body.Velocity()
		body.SetVelocity(v.X, cp.Clamp(v.Y, -FallVelocity, FallVelocity))
	}
}
