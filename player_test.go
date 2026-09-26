package fam

import (
	"math"
	"testing"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/jakecoffman/cp/v2"
	"github.com/jakecoffman/fam/eng"
)

func TestJumpBoost(t *testing.T) {
	g := newTestGame()
	p := g.addPlayer(-1)
	p.grounded = true
	g.Keys[glfw.KeySpace] = true
	p.Update(g, eng.PhysicsDt)
	wantBoost := JumpBoostHeight/math.Sqrt(2*JumpHeight*Gravity) - eng.PhysicsDt
	if math.Abs(p.remainingBoost-wantBoost) > 1e-9 {
		t.Fatalf("boost duration = %v, want %v", p.remainingBoost, wantBoost)
	}
	velocity := p.Velocity().Y
	update := playerUpdateVelocity(g, p)
	update(p.Body, cp.Vector{Y: Gravity}, 1, eng.PhysicsDt)
	if p.Velocity().Y != velocity {
		t.Error("gravity applied while holding boosted jump")
	}
	delete(g.Keys, glfw.KeySpace)
	p.Update(g, eng.PhysicsDt)
	update(p.Body, cp.Vector{Y: Gravity}, 1, eng.PhysicsDt)
	if p.remainingBoost != 0 || p.Velocity().Y <= velocity {
		t.Error("releasing jump did not cancel boost and restore gravity")
	}
	g.Keys[glfw.KeySpace] = true
	p.Update(g, eng.PhysicsDt)
	if p.remainingBoost != 0 {
		t.Error("pressing jump in midair restarted boost")
	}
}

func TestJumpHeight(t *testing.T) {
	for _, test := range []struct {
		name string
		hold bool
		want float64
	}{
		{"tap", false, 250},
		{"hold", true, 500},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := newTestGame()
			g.Space.SetGravity(cp.Vector{Y: Gravity})
			start := cp.Vector{X: worldWidth / 2, Y: 800}
			p := NewPlayer(start, playerRadius, g)
			p.Joystick = -1
			p.grounded = true
			g.Keys[glfw.KeySpace] = true
			height := 0.0
			reachedApex := false
			for i := range 240 {
				if i == 1 && !test.hold {
					delete(g.Keys, glfw.KeySpace)
				}
				p.Update(g, eng.PhysicsDt)
				g.Space.Step(eng.PhysicsDt)
				height = math.Max(height, start.Y-p.Position().Y)
				if p.Velocity().Y >= 0 {
					reachedApex = true
					break
				}
			}
			// Position integration and the shortest tap each contribute up to one step.
			tolerance := 2 * math.Sqrt(2*JumpHeight*Gravity) * eng.PhysicsDt
			if !reachedApex || math.Abs(height-test.want) > tolerance {
				t.Errorf("jump height = %.2f, want %.2f +/- %.2f (reached apex: %v)", height, test.want, tolerance, reachedApex)
			}
		})
	}
}

func TestJumpBoostExpiresAndStopsAtCeiling(t *testing.T) {
	t.Run("expires", func(t *testing.T) {
		g := newTestGame()
		p := g.addPlayer(-1)
		p.grounded = true
		g.Keys[glfw.KeySpace] = true
		p.Update(g, eng.PhysicsDt)
		p.grounded = false
		for range 120 {
			p.Update(g, eng.PhysicsDt)
		}
		before := p.Velocity().Y
		playerUpdateVelocity(g, p)(p.Body, cp.Vector{Y: Gravity}, 1, eng.PhysicsDt)
		if p.remainingBoost != 0 || p.Velocity().Y <= before {
			t.Error("holding jump did not exhaust boost and restore gravity")
		}
	})
	t.Run("ceiling", func(t *testing.T) {
		g := newTestGame()
		p := NewPlayer(cp.Vector{Y: 24}, playerRadius, g)
		p.jumpHeld = true
		p.remainingBoost = 0.5
		g.Space.AddShape(cp.NewSegment(g.Space.StaticBody, cp.Vector{X: -100}, cp.Vector{X: 100}, 1))
		g.Space.Step(eng.PhysicsDt)
		if p.remainingBoost != 0 || p.grounded {
			t.Error("ceiling contact did not cancel boost without grounding player")
		}
	})
}
