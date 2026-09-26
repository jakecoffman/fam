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
	if !p.boosting {
		t.Fatal("jump did not start boost")
	}
	velocity := p.Velocity().Y
	update := playerUpdateVelocity(g, p)
	update(p.Body, cp.Vector{Y: Gravity}, 1, eng.PhysicsDt)
	wantVelocity := velocity + Gravity*JumpHeight/(JumpHeight+JumpBoostHeight)*eng.PhysicsDt
	if math.Abs(p.Velocity().Y-wantVelocity) > 1e-9 {
		t.Errorf("held-jump velocity = %v, want %v", p.Velocity().Y, wantVelocity)
	}
	delete(g.Keys, glfw.KeySpace)
	p.Update(g, eng.PhysicsDt)
	velocity = p.Velocity().Y
	update(p.Body, cp.Vector{Y: Gravity}, 1, eng.PhysicsDt)
	if p.boosting || math.Abs(p.Velocity().Y-(velocity+Gravity*eng.PhysicsDt)) > 1e-9 {
		t.Error("releasing jump did not cancel boost and restore gravity")
	}
	g.Keys[glfw.KeySpace] = true
	p.Update(g, eng.PhysicsDt)
	if p.boosting {
		t.Error("pressing jump in midair restarted boost")
	}
}

func TestJumpHeight(t *testing.T) {
	heights := make(map[string]float64)
	for _, test := range []struct {
		name         string
		releaseAfter float64
		want         float64
	}{
		{"tap", eng.PhysicsDt, 125},
		{"100ms tap", 0.1, 146.35},
		{"hold", math.Inf(1), 187.5},
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
				if float64(i)*eng.PhysicsDt >= test.releaseAfter {
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
			heights[test.name] = height
			t.Logf("jump height: %.2f", height)
		})
	}
	if heights["100ms tap"] > 0.8*heights["hold"] {
		t.Errorf("a 100ms tap (%.2f) should be at least 20%% lower than a held jump (%.2f)", heights["100ms tap"], heights["hold"])
	}
}

func TestJumpBoostStopsAtApexAndCeiling(t *testing.T) {
	t.Run("apex", func(t *testing.T) {
		g := newTestGame()
		p := g.addPlayer(-1)
		p.boosting = true
		p.jumpHeld = true
		p.SetVelocity(0, 1)
		before := p.Velocity().Y
		playerUpdateVelocity(g, p)(p.Body, cp.Vector{Y: Gravity}, 1, eng.PhysicsDt)
		if p.boosting || math.Abs(p.Velocity().Y-(before+Gravity*eng.PhysicsDt)) > 1e-9 {
			t.Error("holding jump while falling did not restore full gravity")
		}
	})
	t.Run("ceiling", func(t *testing.T) {
		g := newTestGame()
		p := NewPlayer(cp.Vector{Y: 24}, playerRadius, g)
		p.jumpHeld = true
		p.boosting = true
		p.SetVelocity(0, -100)
		g.Space.AddShape(cp.NewSegment(g.Space.StaticBody, cp.Vector{X: -100}, cp.Vector{X: 100}, 1))
		g.Space.Step(eng.PhysicsDt)
		if p.boosting || p.grounded {
			t.Error("ceiling contact did not cancel boost without grounding player")
		}
	})
}
