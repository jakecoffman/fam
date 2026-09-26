package fam

import (
	"math"
	"testing"

	"github.com/jakecoffman/cp/v2"
	"github.com/jakecoffman/fam/eng"
)

func testBanana(g *Game, pos cp.Vector) *Banana {
	body := g.Space.AddBody(cp.NewBody(1, 1))
	body.SetPosition(pos)
	shape := g.Space.AddShape(cp.NewCircle(body, 10, cp.Vector{}))
	banana := &Banana{Object: &eng.Object{Body: body, Shape: shape}}
	shape.UserData = banana
	shape.SetCollisionType(collisionBanana)
	return banana
}

func enableFruitCollisions(g *Game) {
	handler := g.Space.NewCollisionHandler(collisionBanana, collisionPlayer)
	handler.PreSolveFunc = BananaPreSolve
	handler.UserData = g
}

func TestConsumedFruitCompactsOnceAndClearsTail(t *testing.T) {
	g := newTestGame()
	enableFruitCollisions(g)
	player := NewPlayer(cp.Vector{X: 100, Y: 100}, playerRadius, g)
	player.Joystick = -1
	g.Players = append(g.Players, player)
	liveA := testBanana(g, cp.Vector{X: 300, Y: 300})
	eatenA := testBanana(g, cp.Vector{X: 100, Y: 100})
	liveB := testBanana(g, cp.Vector{X: 500, Y: 500})
	eatenB := testBanana(g, cp.Vector{X: 110, Y: 100})
	g.Bananas = []*Banana{liveA, eatenA, liveB, eatenB}
	backing := g.Bananas
	shapeA, shapeB := eatenA.Shape, eatenB.Shape
	bodyA, bodyB := eatenA.Body, eatenB.Body

	g.Update(eng.PhysicsDt)

	if len(g.Bananas) != 2 || g.Bananas[0] != liveA || g.Bananas[1] != liveB {
		t.Fatal("compaction did not preserve live fruit order")
	}
	if backing[2] != nil || backing[3] != nil {
		t.Error("discarded slice slots still retain consumed fruit")
	}
	if g.Space.ContainsShape(shapeA) || g.Space.ContainsShape(shapeB) ||
		g.Space.ContainsBody(bodyA) || g.Space.ContainsBody(bodyB) {
		t.Error("consumed fruit remains in physics space")
	}
	if eatenA.Body != nil || eatenB.Body != nil || g.bananasConsumed {
		t.Error("consumption cleanup was incomplete")
	}
	g.Update(eng.PhysicsDt)
	if len(g.Bananas) != 2 {
		t.Error("cleanup removed live fruit on the next step")
	}
}

func TestFruitIsConsumedOnlyOnce(t *testing.T) {
	g := newTestGame()
	enableFruitCollisions(g)
	for range 2 {
		player := NewPlayer(cp.Vector{X: 100, Y: 100}, playerRadius, g)
		player.Joystick = -1
		g.Players = append(g.Players, player)
	}
	g.Bananas = []*Banana{testBanana(g, cp.Vector{X: 100, Y: 100})}
	backing := g.Bananas
	g.Update(eng.PhysicsDt)
	if len(g.Bananas) != 0 || backing[0] != nil {
		t.Error("last consumed fruit remains referenced")
	}
	totalRadius := g.Players[0].Circle.Radius() + g.Players[1].Circle.Radius()
	if math.Abs(totalRadius-playerRadius*2.1) > 1e-9 {
		t.Errorf("one fruit changed combined player radius to %v, want %v", totalRadius, playerRadius*2.1)
	}
}
