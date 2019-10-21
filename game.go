package fam

import (
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/cp/examples"
	"log"
)

type Game struct {
	state         int
	Keys          [1024]bool
	Width, Height float64
	vsync         int

	Space *cp.Space

	Players []*Ball

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

	joys := []glfw.Joystick{glfw.Joystick1, glfw.Joystick2, glfw.Joystick3, glfw.Joystick4}
	colors := []mgl32.Vec3{{0, 1, 0}, {0, 0, 1}, DefaultColor, {1, 0, 0}}
	g.Players = []*Ball{}
	center := cp.Vector{g.Width / 2, g.Height / 2}
	for i, joy := range joys {
		if !glfw.JoystickPresent(joy) {
			break
		}
		g.Players = append(g.Players, NewBall(center, ballRadius, g.Texture("face"), g.Space))
		g.Players[i].Color = colors[i]
		g.Players[i].Joystick = joy
	}

	glfw.SetJoystickCallback(func(joy, event int) {
		if glfw.MonitorEvent(event) == glfw.Connected {
			if joy+1 <= len(g.Players) {
				log.Println("Joystick reconnected", joy)
				return
			}
			log.Println("Joystick connected", joy)
			i := len(g.Players)
			g.Players = append(g.Players, NewBall(center, ballRadius, g.Texture("face"), g.Space))
			g.Players[i].Color = colors[i]
			g.Players[i].Joystick = glfw.Joystick(joy)
		} else {
			log.Println("Joystick disconnected", joy)
		}
	})

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
		if g.Keys[glfw.KeySpace] {
			i := len(g.Players)
			g.Players = append(g.Players, NewBall(center, ballRadius, g.Texture("face"), g.Space))
			g.Players[i].Color = colors[i]
			g.Players[i].Joystick = glfw.Joystick(-1)
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
	for i := range g.Players {
		g.Players[i].processInput(g, dt)
	}
	g.Space.Step(dt)
	//ball := g.Ball
	//g.ParticleGenerator.Update(dt, ball.Position(), ball.Velocity(), 2, mgl32.Vec2{g.Ball.Radius() / 2, g.Ball.Radius() / 2})
}

func (g *Game) Render(alpha float64) {
	//if g.state == stateActive {
	//g.SpriteRenderer.DrawSprite(g.Texture("background"), Vec2(0, 0), mgl32.Vec2{float32(g.Width), float32(g.Height)}, 0, DefaultColor)
	//g.ParticleGenerator.Draw()
	for i := range g.Players {
		g.Players[i].Draw(g.SpriteRenderer, &g.LastBallPosition, alpha)
	}
	//}
	if len(g.Players) == 0 {
		g.TextRenderer.Print("Connect controllers or press SPACE to use keyboard", g.Width/2-250, g.Height/2, 1)
	}
}

func (g *Game) Close() {
	g.Clear()
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
