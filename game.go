package fam

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
)

var GrabbableMaskBit uint = 1 << 31

var GrabFilter = cp.ShapeFilter{
	cp.NO_GROUP, GrabbableMaskBit, GrabbableMaskBit,
}
var NotGrabbableFilter = cp.ShapeFilter{
	cp.NO_GROUP, ^GrabbableMaskBit, ^GrabbableMaskBit,
}

var PlayerMaskBit uint = 1 << 30

var PlayerFilter = cp.ShapeFilter{
	cp.NO_GROUP, PlayerMaskBit, PlayerMaskBit,
}

var NotPlayerFilter = cp.ShapeFilter{
	cp.NO_GROUP, ^PlayerMaskBit, ^PlayerMaskBit,
}


const (
	_ = iota
	collisionPlayer
	collisionBanana
	collisionBomb
)

const (
	actionBanana = "banana"
	actionBomb   = "bomb"
)

type Game struct {
	state      int
	Keys       [1024]bool
	vsync      bool
	fullscreen bool
	window     *OpenGlWindow
	gui        *Gui

	projection mgl32.Mat4

	// mouse stuff
	mouse                 cp.Vector
	mouseBody             *cp.Body
	mouseJoint            *cp.Constraint
	rightDown, rightClick bool
	lmbAction, rmbAction  string

	Space *cp.Space

	Players []*Player
	Bananas []*Banana
	Bombs   []*Bomb

	*ResourceManager
	ParticleGenerator *ParticleGenerator
	SpriteRenderer    *SpriteRenderer
	CPRenderer        *CPRenderer
	TextRenderer      *TextRenderer

	shouldRenderCp bool
}

var colors = []mgl32.Vec3{
	{0, 1, 0}, {0, 0, 1}, DefaultColor, {1, 0, 0},
	{.5, .5, 0}, {0, .5, .5}, {.5, 0, .5}, {.1, .1, .1},
}

// Game state
const (
	stateActive = iota
	statePause
)

var (
	playerVelocity = 11250.0
	playerRadius   = 25.0
)

func (g *Game) New(openGlWindow *OpenGlWindow) {
	g.vsync = true
	g.window = openGlWindow
	g.gui = NewGui(g)
	g.Keys = [1024]bool{}
	g.mouseBody = cp.NewKinematicBody()
	g.lmbAction = actionBanana
	g.rmbAction = actionBomb

	g.ResourceManager = NewResourceManager()

	g.LoadShader("shaders/main.vs.glsl", "shaders/main.fs.glsl", "sprite")
	g.LoadShader("shaders/particle.vs.glsl", "shaders/particle.fs.glsl", "particle")
	g.LoadShader("shaders/cp.vs.glsl", "shaders/cp.fs.glsl", "cp")

	w, h := float64(openGlWindow.Width), float64(openGlWindow.Height)
	center := cp.Vector{w / 2, h / 2}

	g.projection = mgl32.Ortho(0, float32(w), float32(h), 0, -1, 1)
	g.Shader("sprite").Use().SetInt("sprite", 0).SetMat4("projection", g.projection)
	g.Shader("particle").Use().SetInt("sprite", 0).SetMat4("projection", g.projection)
	g.CPRenderer = NewCPRenderer(g.Shader("cp"), g.projection)
	g.SpriteRenderer = NewSpriteRenderer(g.Shader("sprite"))

	shader := g.LoadShader("shaders/text.vs.glsl", "shaders/text.fs.glsl", "text")
	g.TextRenderer = NewTextRenderer(shader, float32(w), float32(h), "fonts/Roboto-Light.ttf", 24)
	g.TextRenderer.SetColor(1, 1, 1, 1)

	// Load all textures by name
	_ = filepath.Walk("textures", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}
		if info.IsDir() {
			return nil
		}
		log.Println("Loading", info.Name())
		g.LoadTexture(fmt.Sprintf("textures/%v", info.Name()), strings.TrimSuffix(info.Name(), filepath.Ext(info.Name())))
		return nil
	})

	g.ParticleGenerator = NewParticleGenerator(g.Shader("particle"), g.Texture("particle"), 500)

	g.reset()

	glfw.SetJoystickCallback(func(joy, event int) {
		if glfw.MonitorEvent(event) == glfw.Connected {
			if joy+1 <= len(g.Players) {
				log.Println("Joystick reconnected", joy)
				return
			}
			log.Println("Joystick connected", joy)
			i := len(g.Players)
			pos := cp.Vector{center.X + rand.Float64()*10, center.Y + rand.Float64()*10}
			g.Players = append(g.Players, NewPlayer(pos, playerRadius, g.Texture("face"), g.Space))
			g.Players[i].Color = colors[i]
			g.Players[i].Joystick = glfw.Joystick(joy)
		} else {
			log.Println("Joystick disconnected", joy)
		}
	})

	g.Players = []*Player{}
	for i := 0; i < 16; i++ {
		joy := glfw.Joystick(i)
		if !glfw.JoystickPresent(joy) {
			break
		}
		g.Players = append(g.Players, NewPlayer(center, playerRadius, g.Texture("face"), g.Space))
		g.Players[i].Color = colors[i]
		g.Players[i].Joystick = joy
	}

	g.state = stateActive

	openGlWindow.SetCursorPosCallback(func(w *glfw.Window, xpos float64, ypos float64) {
		ww, wh := w.GetSize()
		g.mouse = g.MouseToSpace(xpos, ypos, ww, wh)
	})

	openGlWindow.SetKeyCallback(func(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if key == glfw.KeyEscape && action == glfw.Press {
			if g.state == stateActive {
				g.pause()
			} else {
				g.unpause()
			}
		}
		if g.Keys[glfw.KeyF] {
			g.fullscreen = !g.fullscreen
			openGlWindow.SetFullscreen(g.fullscreen)
		}
		if g.Keys[glfw.KeySpace] {
			i := len(g.Players)
			pos := cp.Vector{center.X + rand.Float64()*10, center.Y + rand.Float64()*10}
			g.Players = append(g.Players, NewPlayer(pos, playerRadius, g.Texture("face"), g.Space))
			g.Players[i].Color = colors[i%len(colors)]
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

	openGlWindow.SetMouseButtonCallback(func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
		if g.state != stateActive {
			return
		}
		if button == glfw.MouseButton1 {
			if action == glfw.Press {
				// give the mouse click a little radius to make it easier to click small shapes.
				radius := 5.0

				info := g.Space.PointQueryNearest(g.mouse, radius, GrabFilter)

				if info.Shape != nil && info.Shape.Body().Mass() < cp.INFINITY {
					var nearest cp.Vector
					if info.Distance > 0 {
						nearest = info.Point
					} else {
						nearest = g.mouse
					}

					body := info.Shape.Body()
					g.mouseJoint = cp.NewPivotJoint2(g.mouseBody, body, cp.Vector{}, body.WorldToLocal(nearest))
					g.mouseJoint.SetMaxForce(50000)
					g.mouseJoint.SetErrorBias(math.Pow(1.0-0.15, 60.0))
					g.Space.AddConstraint(g.mouseJoint)
				} else {
					if g.lmbAction == actionBanana {
						g.Bananas = append(g.Bananas, NewBanana(g.mouse, 10, g.Texture("banana"), g.Space))
					} else if g.lmbAction == actionBomb {
						g.Bombs = append(g.Bombs, NewBomb(g.mouse, 20, g.Texture("bomb"), g.Texture("pow"), g.Space))
					}
				}
			} else if g.mouseJoint != nil {
				g.Space.RemoveConstraint(g.mouseJoint)
				g.mouseJoint = nil
			}
		} else if button == glfw.MouseButton2 {
			g.rightDown = action == glfw.Press
			g.rightClick = g.rightDown

			if action == glfw.Press {
				if g.rmbAction == actionBanana {
					g.Bananas = append(g.Bananas, NewBanana(g.mouse, 10, g.Texture("banana"), g.Space))
				} else if g.rmbAction == actionBomb {
					g.Bombs = append(g.Bombs, NewBomb(g.mouse, 20, g.Texture("bomb"), g.Texture("pow"), g.Space))
				}
			}
		}
	})
}

func (g *Game) Update(dt float64) {
	if g.state == statePause {
		return
	}
	// update mouse body
	newPoint := g.mouseBody.Position().Lerp(g.mouse, 0.25)
	g.mouseBody.SetVelocityVector(newPoint.Sub(g.mouseBody.Position()).Mult(60.0))
	g.mouseBody.SetPosition(newPoint)

	for i := range g.Bombs {
		g.Bombs[i].Update(g, dt)
	}
	for i := range g.Bananas {
		g.Bananas[i].Update(g, dt)
	}
	for i := range g.Players {
		g.Players[i].Update(g, dt)
	}

	g.Space.Step(dt)
	//ball := g.Player
	//g.ParticleGenerator.Update(dt, ball.Position(), ball.Velocity(), 2, mgl32.Vec2{g.Player.Radius() / 2, g.Player.Radius() / 2})
}

func (g *Game) Render(alpha float64) {
	if g.window.UpdateViewport {
		g.window.UpdateViewport = false
		g.window.ViewportWidth, g.window.ViewPortHeight = g.window.GetFramebufferSize()
		gl.Viewport(0, 0, int32(g.window.ViewportWidth), int32(g.window.ViewPortHeight))
		log.Printf("update viewport %#v\n", g.window)
	}

	if g.shouldRenderCp {
		g.CPRenderer.DrawSpace(g.Space)
	}

	//if g.state == stateActive {
	//g.SpriteRenderer.DrawSprite(g.Texture("background"), Vec2(0, 0), mgl32.Vec2{float32(g.window.Width), float32(g.window.Height)}, 0, DefaultColor)
	//g.ParticleGenerator.Draw()
	for i := range g.Bananas {
		g.Bananas[i].Draw(g.SpriteRenderer, alpha)
	}
	for i := range g.Bombs {
		g.Bombs[i].Draw(g.SpriteRenderer, alpha)
	}
	for i := range g.Players {
		g.Players[i].Draw(g.SpriteRenderer, alpha)
	}
	//}

	if len(g.Players) == 0 {
		g.TextRenderer.Print("Connect controllers or press SPACE to use keyboard", float64(g.window.Width)/2.-250., float64(g.window.Height)/2., 1)
	}

	//g.SpriteRenderer.DrawSprite(g.Texture("banana"), V(g.mouse), mgl32.Vec2{100, 100}, 0, mgl32.Vec3{1, 0, 0})

	if g.state == statePause {
		g.gui.Render()
	}
}

func (g *Game) Close() {
	g.gui.Destroy()
	g.Clear()
}

func (g *Game) pause() {
	g.state = statePause
}

func (g *Game) unpause() {
	g.state = stateActive
}

func (g *Game) reset() {
	g.Space = cp.NewSpace()
	g.Space.SetGravity(cp.Vector{0, 0})

	bananaCollisionHandler := g.Space.NewCollisionHandler(collisionBanana, collisionPlayer)
	bananaCollisionHandler.PreSolveFunc = BananaPreSolve
	bananaCollisionHandler.UserData = g

	bombCollisionHandler := g.Space.NewCollisionHandler(collisionBomb, collisionPlayer)
	bombCollisionHandler.PreSolveFunc = BombPreSolve

	w, h := float64(g.window.Width), float64(g.window.Height)
	center := cp.Vector{w / 2, h / 2}

	const offset = 0
	sides := []cp.Vector{
		{0 - offset, 0 - offset}, {w - offset, 0 - offset},
		{w - offset, 0 - offset}, {w - offset, h - offset},
		{w - offset, h - offset}, {0 - offset, h - offset},
		{0 - offset, h - offset}, {0 - offset, 0 - offset},
	}

	for i := 0; i < len(sides); i += 2 {
		var seg *cp.Shape
		seg = g.Space.AddShape(cp.NewSegment(g.Space.StaticBody, sides[i], sides[i+1], 10))
		seg.SetElasticity(1)
		seg.SetFriction(1)
		seg.SetFilter(NotGrabbableFilter)
	}

	for _, p := range g.Players {
		pos := cp.Vector{center.X + rand.Float64()*10, center.Y + rand.Float64()*10}
		p.Reset(pos, playerRadius, g.Space)
	}
	g.Bananas = []*Banana{}
	g.Bombs = []*Bomb{}
}

func (g *Game) MouseToSpace(x, y float64, ww, wh int) cp.Vector {
	model := mgl32.Translate3D(0, 0, 0)
	obj, err := mgl32.UnProject(mgl32.Vec3{float32(x), float32(float64(wh) - y), 0}, model, g.projection, 0, 0, ww, wh)
	if err != nil {
		panic(err)
	}

	return cp.Vector{float64(obj.X()), float64(obj.Y())}
}
