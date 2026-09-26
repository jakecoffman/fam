package fam

import (
	"testing"

	"github.com/jakecoffman/cp/v2"
	"github.com/jakecoffman/fam/eng"
)

func TestBombTransitions(t *testing.T) {
	g := newTestGame()
	bomb := NewBomb(cp.Vector{X: 100, Y: 100}, 20, g.Space)
	bomb.Update(g, 5.01)
	if bomb.state != bombStateBoom || bomb.Circle.Radius() != 200 {
		t.Fatal("bomb did not enter explosion at ten times its original radius")
	}
	bomb.Update(g, 0.2)
	for range 7 {
		if bomb.state != bombStateSpent || bomb.Circle.Radius() != 20 {
			t.Fatalf("spent bomb state/radius = %v/%v, want %v/20", bomb.state, bomb.Circle.Radius(), bombStateSpent)
		}
		bomb.Update(g, 0.1)
	}
	bomb.Update(g, 0.2)
	if bomb.state != bombStateGone || g.Space.ContainsShape(bomb.Shape) || g.Space.ContainsBody(bomb.Body) {
		t.Error("expired bomb remains in physics space")
	}
	bomb.Update(g, 1)
}

func TestOnlyExplodingBombDeflatesPlayers(t *testing.T) {
	for _, test := range []struct {
		name   string
		time   float64
		radius float64
	}{
		{"exploding", 5.01, playerRadius},
		{"spent", 5.21, 100},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := newTestGame()
			g.Space.NewWildcardCollisionHandler(collisionBomb).PreSolveFunc = BombPreSolve
			player := NewPlayer(cp.Vector{X: 100, Y: 100}, 100, g)
			bomb := NewBomb(cp.Vector{X: 120, Y: 80}, 20, g.Space)
			bomb.Update(g, test.time)
			g.Space.Step(eng.PhysicsDt)
			if player.Circle.Radius() != test.radius {
				t.Errorf("player radius = %v, want %v", player.Circle.Radius(), test.radius)
			}
		})
	}
}
