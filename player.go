package fam

import (
	"math"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp/v2"
	"github.com/jakecoffman/fam/eng"
)

type Player struct {
	Color mgl32.Vec3

	*eng.Object
	Circle *cp.Circle

	Joystick glfw.Joystick

	remainingBoost          float64
	grounded, lastJumpState bool

	// inputX and jumpHeld are polled once per frame in Update and consumed by
	// the velocity callback (which may run multiple times per Step).
	inputX   float64
	jumpHeld bool
}

func NewPlayer(pos cp.Vector, radius float64, g *Game) *Player {
	p := &Player{
		Object: &eng.Object{},
		Color:  mgl32.Vec3{1, 1, 1},
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
	p.Object.Update(g.Space, dt, worldWidth, worldHeight)

	// Poll input once per frame and stash results so that playerUpdateVelocity
	// (which Chipmunk may invoke multiple times per Step) sees consistent state.
	const deadzone = 0.15
	if p.Joystick > -1 {
		axes := glfw.GetJoystickAxes(p.Joystick)
		if len(axes) >= 1 {
			raw := float64(axes[0])
			if math.Abs(raw) < deadzone {
				p.inputX = 0
			} else {
				p.inputX = (raw - math.Copysign(deadzone, raw)) / (1 - deadzone)
			}
		} else {
			p.inputX = 0
		}

		buttonBytes := glfw.GetJoystickButtons(p.Joystick)
		if len(buttonBytes) > 0 {
			p.jumpHeld = glfw.Action(buttonBytes[0]) == glfw.Press
		} else {
			p.jumpHeld = false
		}
	} else {
		if g.Keys[glfw.KeyA] || g.Keys[glfw.KeyLeft] {
			p.inputX = -1
		} else if g.Keys[glfw.KeyD] || g.Keys[glfw.KeyRight] {
			p.inputX = 1
		} else {
			p.inputX = 0
		}
		p.jumpHeld = g.Keys[glfw.KeySpace]
	}

	// If the jump key was just pressed this frame, jump!
	if p.jumpHeld && !p.lastJumpState && p.grounded {
		jumpV := -math.Sqrt(2.0 * JumpHeight * Gravity)
		p.SetVelocityVector(p.Velocity().Add(cp.Vector{0, jumpV}))

		p.remainingBoost = JumpBoostHeight / jumpV
	}
	p.remainingBoost -= dt
	p.lastJumpState = p.jumpHeld
}

func (p *Player) Draw(g *Game, alpha float64) {
	g.SpriteRenderer.DrawSprite(
		g.Texture("face"),
		p.SmoothPos(alpha),
		p.Size().Mul(1.1), // increase 10% to better fit hitbox
		p.SmoothAngle(alpha),
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
)

func playerUpdateVelocity(g *Game, p *Player) func(*cp.Body, cp.Vector, float64, float64) {
	return func(body *cp.Body, gravity cp.Vector, damping, dt float64) {
		// Use pre-polled input (set once per frame in Player.Update).
		x := p.inputX
		jumpState := p.jumpHeld

		// Grab the grounding normal from last frame.
		// Only count normals pointing sufficiently upward (n.Y > 0.7) to avoid
		// wall-sticking letting the player jump off steep walls.
		groundNormal := cp.Vector{}
		body.EachArbiter(func(arb *cp.Arbiter) {
			n := arb.Normal()
			if n.Y > 0.7 && n.Y > groundNormal.Y {
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

		// Apply air control if not grounded
		v := p.Velocity()
		if !p.grounded {
			p.SetVelocity(cp.LerpConst(v.X, targetVx, PlayerAirAccel*dt), v.Y)
		} else {
			p.SetVelocity(targetVx, v.Y)
		}

		// Clamp only the downward (falling) velocity; upward jumps are not clamped.
		v = body.Velocity()
		if v.Y > FallVelocity {
			body.SetVelocity(v.X, FallVelocity)
		}
	}
}
