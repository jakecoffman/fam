package fam

import (
	"math"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/eng"
)

type PlayerSystem struct {
	*System
	game *Game
}

type Player struct {
	Object

	Color   mgl32.Vec3
	Circle *cp.Circle

	Joystick glfw.Joystick

	remainingBoost float64
	grounded, lastJumpState bool
}

func NewPlayerSystem(game *Game) *PlayerSystem {
	const maxPlayers = 16
	return &PlayerSystem{
		System: NewSystem(Player{}, maxPlayers),
		game: game,
	}
}

const playerRadius = 25.0

func (s *PlayerSystem) Add(pos cp.Vector, color mgl32.Vec3, joystick glfw.Joystick) *Player {
	p := s.System.Add().(*Player)
	p.Color = color
	p.Joystick = joystick

	p.Body = cp.NewBody(1, cp.MomentForCircle(1, playerRadius, playerRadius, cp.Vector{0, 0}))
	p.Body.SetVelocityUpdateFunc(playerUpdateVelocity(s.game, p))

	p.Shape = cp.NewCircle(p.Body, playerRadius, cp.Vector{0, 0})
	p.Shape.SetElasticity(0)
	p.Shape.SetFriction(1)

	p.Shape.SetFilter(cp.NewShapeFilter(uint(eng.NextEntityID()), PlayerMaskBit, PlayerMaskBit))
	p.Shape.SetCollisionType(collisionPlayer)
	p.Shape.UserData = p

	p.Circle = p.Shape.Class.(*cp.Circle)
	p.Body.SetPosition(pos)

	s.game.Space.AddBody(p.Body)
	s.game.Space.AddShape(p.Shape)
	return p
}

func (s *PlayerSystem) Update(dt float64) {
	players := s.pool.([]Player)
	for i := 0; i < s.active; i++ {
		players[i].Update(s.game, dt)
	}
}

func (p *Player) Update(g *Game, dt float64) {
	p.Object.Update(dt)
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

func (s *PlayerSystem) Draw(alpha float64) {
	players := s.pool.([]Player)
	for i := 0; i < s.active; i++ {
		players[i].Draw(s.game, alpha)
	}
}

func (p *Player) Draw(g *Game, alpha float64) {
	g.SpriteRenderer.DrawSprite(
		g.Texture("face"),
		p.SmoothPos(alpha),
		p.Size().Mul(1.1), // increase 10% to better fit hitbox
		p.Angle(),
		p.Color)
}

func (s *PlayerSystem) Remove(id eng.EntityID) {
	s.System.Get(id).(*Player).Remove(s.game.Space)
	s.System.Remove(id)
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
