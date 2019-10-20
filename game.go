package fam

import (
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/cp/examples"
)

type Game struct {
	state         int
	Keys          [1024]bool
	Width, Height float64
	vsync         int

	Space *cp.Space

	Ball *Ball

	// used for slerp
	LastBallPosition cp.Vector

	*ResourceManager
	ParticleGenerator *ParticleGenerator
	SpriteRenderer    *SpriteRenderer
	TextRenderer      *TextRenderer
}

// Game state
const (
	stateActive = iota
	stateMenu
	stateWin
)

var (
	playerSize          = mgl32.Vec2{100, 20}
	playerVelocity      = 11250.0
	initialBallVelocity = Vec2(0, 0)
	ballRadius          = 25.0
)

func (g *Game) New(w, h int, window *glfw.Window) {
	g.vsync = 1
	g.Width = float64(w)
	g.Height = float64(h)
	g.Keys = [1024]bool{}
	g.Space = cp.NewSpace()
	g.Space.SetGravity(cp.Vector{0, 0})
	sides := []cp.Vector{
		{0, 0}, {g.Width, 0},
		{g.Width, 0}, {g.Width, g.Height},
		{g.Width, g.Height}, {0, g.Height},
		{0, g.Height}, {0, 0},
	}

	for i := 0; i < len(sides); i += 2 {
		var seg *cp.Shape
		seg = g.Space.AddShape(cp.NewSegment(g.Space.StaticBody, sides[i], sides[i+1], 10))
		seg.SetElasticity(1)
		seg.SetFriction(1)
		seg.SetFilter(examples.NotGrabbableFilter)
	}

	g.ResourceManager = NewResourceManager()

	width, height := float32(g.Width), float32(g.Height)

	g.LoadShader("shaders/main.vs.glsl", "shaders/main.fs.glsl", "sprite")
	g.LoadShader("shaders/particle.vs.glsl", "shaders/particle.fs.glsl", "particle")

	projection := mgl32.Ortho(0, width, height, 0, -1, 1)
	g.Shader("sprite").Use().
		SetInt("sprite", 0).
		SetMat4("projection", projection)
	g.Shader("particle").Use().
		SetInt("sprite", 0).
		SetMat4("projection", projection)

	g.LoadTexture("textures/background.jpg", "background")
	g.LoadTexture("textures/paddle.png", "paddle")
	g.LoadTexture("textures/particle.png", "particle")
	g.LoadTexture("textures/awesomeface.png", "face")
	g.LoadTexture("textures/banana.png", "banana")

	shader := g.LoadShader("shaders/text.vs.glsl", "shaders/text.fs.glsl", "text")
	g.TextRenderer = NewTextRenderer(shader, width, height, "textures/Roboto-Light.ttf", 24)
	g.TextRenderer.SetColor(1, 1, 1, 1)

	g.ParticleGenerator = NewParticleGenerator(g.Shader("particle"), g.Texture("particle"), 500)
	g.SpriteRenderer = NewSpriteRenderer(g.Shader("sprite"))

	//playerPos := cp.Vector{g.Width/2.0, g.Height/2}
	g.Ball = NewBall(cp.Vector{100, 100}, ballRadius, g.Texture("face"))
	g.Space.AddBody(g.Ball.Body)
	g.Space.AddShape(g.Ball.Shape)
	//g.Ball.Body = cp.NewBody(1, cp.MomentForCircle(1, ballRadius, ballRadius, cp.Vector{0, 0}))

	g.state = stateActive

	window.SetKeyCallback(func(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if key == glfw.KeyEscape && action == glfw.Press {
			window.SetShouldClose(true)
		}
		if g.Keys[glfw.KeyV] {
			if g.vsync == 0 {
				g.vsync = 1
			} else {
				g.vsync = 0
			}
			glfw.SwapInterval(g.vsync)
		}
		// store for continuous application
		if key >= 0 && key < 1024 {
			if action == glfw.Press {
				g.Keys[key] = true
			} else if action == glfw.Release {
				g.Keys[key] = false
			}
		}
	})
}

func (g *Game) Update(dt float64) {
	g.LastBallPosition = g.Ball.Position()
	g.processInput(dt)
	g.Space.Step(dt)
	//ball := g.Ball
	//g.ParticleGenerator.Update(dt, ball.Position(), ball.Velocity(), 2, mgl32.Vec2{g.Ball.Radius() / 2, g.Ball.Radius() / 2})
}

func (g *Game) Render(alpha float64) {
	//if g.state == stateActive {
	g.SpriteRenderer.DrawSprite(g.Texture("background"), Vec2(0, 0), mgl32.Vec2{float32(g.Width), float32(g.Height)}, 0, DefaultColor)
	//g.ParticleGenerator.Draw()
	g.Ball.Draw(g.SpriteRenderer, &g.LastBallPosition, alpha)
	//}
	g.TextRenderer.Print("Hello, world!", 10, 25, 1)
}

func (g *Game) Close() {
	g.Clear()
}

func (g *Game) processInput(dt float64) {
	//if g.state != stateActive {
	//	return
	//}

	velocity := playerVelocity * dt

	force := cp.Vector{}
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

	g.Ball.SetVelocityVector(force)
}

func (g *Game) pause() {
	g.state = stateMenu
}

func (g *Game) unpause() {
	g.state = stateActive
}

type Direction int

const (
	directionUp Direction = iota
	directionRight
	directionDown
	directionLeft
)

func vectorDirection(target mgl32.Vec2) Direction {
	compass := []mgl32.Vec2{
		{0, 1},
		{1, 0},
		{0, -1},
		{-1, 0},
	}
	var max float32 = 0.0
	bestMatch := -1
	for i := 0; i < 4; i++ {
		dotProduct := target.Normalize().Dot(compass[i])
		if dotProduct > max {
			max = dotProduct
			bestMatch = i
		}
	}
	return Direction(bestMatch)
}
