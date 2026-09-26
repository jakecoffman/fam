package fam

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/jakecoffman/cp/v2"
	"github.com/jakecoffman/fam/eng"
)

func newTestGame() *Game {
	return &Game{
		Space:        cp.NewSpace(),
		Keys:         make(map[glfw.Key]bool),
		mouseBody:    cp.NewKinematicBody(),
		WallRenderer: &eng.CPRenderer{},
	}
}

func testLevel(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "level.json")
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadLevelReplacesWalls(t *testing.T) {
	g := newTestGame()
	player := g.addPlayer(-1)
	path := testLevel(t, `[{"A":{"X":10,"Y":10},"B":{"X":100,"Y":10}}]`)
	if err := g.loadLevel(path); err != nil {
		t.Fatal(err)
	}
	for range 3 {
		oldWall := g.Walls[0]
		g.updateWallMesh()
		if g.wallsDirty {
			t.Fatal("wall mesh remains dirty after rebuild")
		}
		if err := g.loadLevel(path); err != nil {
			t.Fatal(err)
		}
		if g.Space.ContainsShape(oldWall.Shape) {
			t.Error("old wall remains in physics space")
		}
		shapes := 0
		g.Space.EachShape(func(*cp.Shape) { shapes++ })
		if shapes != 2 || len(g.Walls) != 1 {
			t.Fatalf("got %d physics shapes and %d walls, want 2 and 1", shapes, len(g.Walls))
		}
		if !g.Space.ContainsBody(player.Body) || !g.Space.ContainsShape(player.Shape) {
			t.Error("loading level removed player")
		}
		if !g.wallsDirty {
			t.Error("loading level did not invalidate wall mesh")
		}
	}

	oldWall := g.Walls[0]
	g.updateWallMesh()
	if err := g.loadLevel(testLevel(t, `invalid`)); err == nil {
		t.Fatal("invalid level accepted")
	}
	if g.Walls[0] != oldWall || !g.Space.ContainsShape(oldWall.Shape) || g.wallsDirty {
		t.Error("invalid load changed the current level")
	}
	if err := g.loadLevel(testLevel(t, `[]`)); err != nil {
		t.Fatal(err)
	}
	if len(g.Walls) != 0 || g.Space.ContainsShape(oldWall.Shape) || !g.wallsDirty {
		t.Error("empty level did not remove old walls and invalidate mesh")
	}
}

func TestWallEditing(t *testing.T) {
	g := newTestGame()
	g.mouse = cp.Vector{X: 100, Y: 100}
	g.handleMouseButton(glfw.MouseButton1, glfw.Press)
	if len(g.Walls) != 0 || g.drawingWallShape == nil {
		t.Fatal("preview should remain separate from committed walls")
	}
	g.mouse = cp.Vector{X: 200, Y: 100}
	g.handleMouseButton(glfw.MouseButton1, glfw.Release)
	if len(g.Walls) != 1 || !g.wallsDirty || g.drawingWallShape != nil {
		t.Fatal("release did not commit wall and invalidate mesh")
	}
	wall := g.Walls[0]
	if !g.Space.ContainsShape(wall.Shape) || !wall.B().Equal(g.mouse) {
		t.Fatal("wall was not committed at the release position")
	}
	g.updateWallMesh()
	g.mouse = cp.Vector{X: 150, Y: 100}
	g.handleMouseButton(glfw.MouseButton2, glfw.Press)
	g.handleMouseButton(glfw.MouseButton2, glfw.Release)
	if len(g.Walls) != 0 || g.Space.ContainsShape(wall.Shape) || !g.wallsDirty {
		t.Error("wall deletion did not update physics and mesh")
	}
	if !g.Space.ContainsBody(g.Space.StaticBody) {
		t.Error("wall deletion detached the shared static body")
	}

	g.mouse = cp.Vector{X: 300, Y: 100}
	g.handleMouseButton(glfw.MouseButton1, glfw.Press)
	g.handleMouseButton(glfw.MouseButton1, glfw.Release)
	if len(g.Walls) != 0 {
		t.Error("click without dragging created a zero-length wall")
	}
}

func TestLoadAndResetCancelMouseInteraction(t *testing.T) {
	g := newTestGame()
	player := g.addPlayer(3)
	g.mouse = player.Position()
	g.handleMouseButton(glfw.MouseButton1, glfw.Press)
	joint := g.mouseJoint
	if joint == nil {
		t.Fatal("failed to grab player")
	}
	if err := g.loadLevel(testLevel(t, `[]`)); err != nil {
		t.Fatal(err)
	}
	if g.mouseJoint != nil || g.Space.ContainsConstraint(joint) {
		t.Error("loading level retained the mouse constraint")
	}
	g.mouse = cp.Vector{X: 100, Y: 100}
	g.handleMouseButton(glfw.MouseButton1, glfw.Press)
	oldSpace := g.Space
	g.reset()
	if g.Space == oldSpace || g.drawingWallShape != nil || g.leftDown != nil {
		t.Error("reset retained old editor state")
	}
	g.handleMouseButton(glfw.MouseButton1, glfw.Release)
	for _, wall := range g.Walls {
		if wall.Body() != g.Space.StaticBody {
			t.Error("wall belongs to the old physics space")
		}
	}
	if len(g.Players) != 1 || g.Players[0] != player || !g.Space.ContainsBody(player.Body) {
		t.Error("reset did not move the controller player to the new space")
	}
}

func TestKeyActionsOnlyOnPress(t *testing.T) {
	g := newTestGame()
	g.handleKey(glfw.KeyEnter, glfw.Press)
	if len(g.Players) != 1 || g.Players[0].Joystick != -1 {
		t.Fatal("Enter did not immediately add a keyboard player")
	}
	g.handleKey(glfw.KeyEnter, glfw.Repeat)
	g.handleKey(glfw.KeyA, glfw.Press)
	g.handleKey(glfw.KeyEnter, glfw.Release)
	if len(g.Players) != 1 || g.Keys[glfw.KeyEnter] || !g.Keys[glfw.KeyA] {
		t.Error("repeat/unrelated/release events changed players or key state")
	}
	g.handleKey(glfw.KeyQ, glfw.Press)
	g.handleKey(glfw.KeyQ, glfw.Repeat)
	g.handleKey(glfw.KeyA, glfw.Release)
	g.handleKey(glfw.KeyQ, glfw.Release)
	if len(g.Bombs) != 1 {
		t.Errorf("got %d bombs for one Q press, want 1", len(g.Bombs))
	}
	g.handleKey(glfw.KeyEscape, glfw.Press)
	for _, key := range []glfw.Key{glfw.KeyEnter, glfw.KeyQ, glfw.KeyE, glfw.KeyF} {
		g.handleKey(key, glfw.Press)
		g.handleKey(key, glfw.Release)
	}
	if len(g.Players) != 1 || len(g.Bombs) != 1 || len(g.Bananas) != 0 || g.fullscreen {
		t.Error("game actions ran while paused")
	}
	g.handleKey(glfw.KeyEscape, glfw.Repeat)
	g.handleKey(glfw.KeyEscape, glfw.Release)
	if g.state != statePause {
		t.Error("Escape repeat/release changed pause state")
	}
	g.handleKey(glfw.KeyEscape, glfw.Press)
	if g.state != stateActive {
		t.Error("Escape press did not resume")
	}
}

func TestJoystickIdentityAndSparseDiscovery(t *testing.T) {
	g := newTestGame()
	keyboard := g.addPlayer(-1)
	g.connectJoystick(0)
	controller := g.Players[1]
	if controller.Joystick != 0 || g.Players[0] != keyboard {
		t.Fatal("keyboard player interfered with controller connection")
	}
	g.connectJoystick(0)
	if len(g.Players) != 2 || g.Players[1] != controller {
		t.Error("reconnection duplicated or replaced a controller player")
	}
	visited := 0
	g.discoverJoysticks(func(joy glfw.Joystick) bool {
		visited++
		return joy == 0 || joy == 5 || joy == 15
	})
	if visited != 16 || len(g.Players) != 4 || g.Players[2].Joystick != 5 || g.Players[3].Joystick != 15 {
		t.Errorf("sparse discovery visited %d slots and found %d players", visited, len(g.Players))
	}
}
