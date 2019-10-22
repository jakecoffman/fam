package fam

import (
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/cp/examples"
	"log"
)

type Game struct {
	state      int
	Keys       [1024]bool
	vsync      int
	fullscreen bool
	window     *OpenGlWindow

	Space *cp.Space

	Players []*Ball

	// used for slerp
	LastBallPosition cp.Vector

	*ResourceManager
	ParticleGenerator *ParticleGenerator
	SpriteRenderer    *SpriteRenderer
	PrimitiveRenderer *PrimitiveRenderer
	CPRenderer        *CPRenderer
	TextRenderer      *TextRenderer
}

// Game state
const (
	stateActive = iota
	stateMenu
	stateWin
)

var (
	playerVelocity = 11250.0
	ballRadius     = 25.0
)

func (g *Game) New(openGlWindow *OpenGlWindow) {
	g.vsync = 1
	g.window = openGlWindow
	g.Keys = [1024]bool{}
	g.Space = cp.NewSpace()
	g.Space.SetGravity(cp.Vector{0, 0})

	w, h := float64(openGlWindow.Width), float64(openGlWindow.Height)
	const offset = 0
	sides := []cp.Vector{
		{0 - offset, 0 - offset}, {w - offset, 0 - offset},
		{w - offset, 0 - offset}, {w - offset, h - offset},
		{w - offset, h - offset}, {0 - offset, h - offset},
		{0 - offset, h - offset}, {0 - offset, 0 - offset},
	}

	for i := 0; i < len(sides); i += 2 {
		var seg *cp.Shape
		seg = g.Space.AddShape(cp.NewSegment(g.Space.StaticBody, sides[i], sides[i+1], 1))
		seg.SetElasticity(1)
		seg.SetFriction(1)
		seg.SetFilter(examples.NotGrabbableFilter)
	}

	g.ResourceManager = NewResourceManager()

	g.LoadShader("shaders/main.vs.glsl", "shaders/main.fs.glsl", "sprite")
	g.LoadShader("shaders/particle.vs.glsl", "shaders/particle.fs.glsl", "particle")
	g.LoadShader("shaders/primitive.vs.glsl", "shaders/primitive.fs.glsl", "primitive")
	g.LoadShader("shaders/cp.vs.glsl", "shaders/cp.fs.glsl", "cp")

	projection := mgl32.Ortho(0, float32(w), float32(h), 0, -1, 1)
	g.Shader("sprite").Use().
		SetInt("sprite", 0).
		SetMat4("projection", projection)
	g.Shader("particle").Use().
		SetInt("sprite", 0).
		SetMat4("projection", projection)
	g.Shader("primitive").Use().
		SetMat4("projection", projection)

	g.LoadTexture("textures/background.jpg", "background")
	g.LoadTexture("textures/paddle.png", "paddle")
	g.LoadTexture("textures/particle.png", "particle")
	g.LoadTexture("textures/awesomeface.png", "face")
	g.LoadTexture("textures/banana.png", "banana")

	shader := g.LoadShader("shaders/text.vs.glsl", "shaders/text.fs.glsl", "text")
	g.TextRenderer = NewTextRenderer(shader, float32(w), float32(h), "textures/Roboto-Light.ttf", 24)
	g.TextRenderer.SetColor(1, 1, 1, 1)

	g.ParticleGenerator = NewParticleGenerator(g.Shader("particle"), g.Texture("particle"), 500)
	g.SpriteRenderer = NewSpriteRenderer(g.Shader("sprite"))
	g.PrimitiveRenderer = NewPrimitiveRenderer(g.Shader("primitive"))
	g.CPRenderer = NewCPRenderer(g.Shader("cp"), projection)

	joys := []glfw.Joystick{glfw.Joystick1, glfw.Joystick2, glfw.Joystick3, glfw.Joystick4}
	colors := []mgl32.Vec3{{0, 1, 0}, {0, 0, 1}, DefaultColor, {1, 0, 0}}
	g.Players = []*Ball{}
	center := cp.Vector{w / 2, h / 2}
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

	openGlWindow.SetKeyCallback(func(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
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
		if g.Keys[glfw.KeyF] {
			g.fullscreen = !g.fullscreen
			openGlWindow.SetFullscreen(g.fullscreen)
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
	//g.SpriteRenderer.DrawSprite(g.Texture("background"), Vec2(0, 0), mgl32.Vec2{float32(g.Width), float32(h)}, 0, DefaultColor)
	//g.ParticleGenerator.Draw()
	for i := range g.Players {
		g.Players[i].Draw(g.SpriteRenderer, &g.LastBallPosition, alpha)
	}
	//}

	//g.PrimitiveRenderer.DrawPrimitive(mgl32.Vec2{100, 100}, mgl32.Vec2{100, 100}, 0, mgl32.Vec3{1, 0, 0})
	//g.CPRenderer.DrawSpace(g.Space, false)

	if len(g.Players) == 0 {
		g.TextRenderer.Print("Connect controllers or press SPACE to use keyboard", float64(g.window.Width)/2.-250., float64(g.window.Height)/2., 1)
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
