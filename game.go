package fam

import (
	"encoding/json"
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
	"github.com/jakecoffman/fam/eng"
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

const wallWidth = 10

type Game struct {
	state      int
	Keys       [1024]bool
	vsync      bool
	fullscreen bool
	window     *eng.OpenGlWindow
	gui        *Gui

	projection mgl32.Mat4

	// mouse stuff
	mouse      cp.Vector
	mouseBody  *cp.Body
	mouseJoint *cp.Constraint

	leftDown  *cp.Vector
	rightDown *cp.Vector

	drawingWallShape *cp.Segment

	Space *cp.Space

	Players []*Player
	Bananas []*Banana
	Bombs   []*Bomb
	Walls   []*cp.Segment

	*eng.ResourceManager
	ParticleGenerator *eng.ParticleGenerator
	SpriteRenderer    *eng.SpriteRenderer
	CPRenderer        *eng.CPRenderer
	TextRenderer      *eng.TextRenderer

	shouldRenderCp bool

	level string
}

const (
	worldWidth  = 1920
	worldHeight = 1080
)

// Game state
const (
	stateActive = iota
	statePause
)

const (
	playerVelocity = 11250.0
	playerRadius   = 25.0
)

func (g *Game) New(openGlWindow *eng.OpenGlWindow) {
	g.vsync = true
	g.window = openGlWindow
	g.gui = NewGui(g)
	g.Keys = [1024]bool{}
	g.mouseBody = cp.NewKinematicBody()

	g.ResourceManager = eng.NewResourceManager()

	g.LoadShader("assets/shaders/main.vs.glsl", "assets/shaders/main.fs.glsl", "sprite")
	g.LoadShader("assets/shaders/particle.vs.glsl", "assets/shaders/particle.fs.glsl", "particle")
	g.LoadShader("assets/shaders/cp.vs.glsl", "assets/shaders/cp.fs.glsl", "cp")
	g.LoadShader("assets/shaders/text.vs.glsl", "assets/shaders/text.fs.glsl", "text")

	center := cp.Vector{worldWidth / 2, worldHeight / 2}

	g.projection = mgl32.Ortho(0, worldWidth, worldHeight, 0, -1, 1)
	g.Shader("sprite").Use().SetInt("sprite", 0).SetMat4("projection", g.projection)
	g.Shader("particle").Use().SetInt("sprite", 0).SetMat4("projection", g.projection)
	g.CPRenderer = eng.NewCPRenderer(g.Shader("cp"), g.projection)
	g.SpriteRenderer = eng.NewSpriteRenderer(g.Shader("sprite"))
	g.TextRenderer = eng.NewTextRenderer(g.Shader("text"), float32(openGlWindow.Width), float32(openGlWindow.Height), "assets/fonts/Roboto-Light.ttf", 24)
	g.TextRenderer.SetColor(1, 1, 1, 1)

	// Load all textures by name
	_ = filepath.Walk("assets/textures", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}
		if info.IsDir() {
			return nil
		}
		log.Println("Loading", info.Name())
		g.LoadTexture(fmt.Sprintf("assets/textures/%v", info.Name()), strings.TrimSuffix(info.Name(), filepath.Ext(info.Name())))
		return nil
	})

	g.ParticleGenerator = eng.NewParticleGenerator(g.Shader("particle"), g.Texture("particle"), 500)

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
			g.Players = append(g.Players, NewPlayer(pos, playerRadius, g))
			g.Players[i].Color = eng.NextColor()
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
		g.Players = append(g.Players, NewPlayer(center, playerRadius, g))
		g.Players[i].Color = eng.NextColor()
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
		if g.Keys[glfw.KeyE] {
			g.Bananas = append(g.Bananas, NewBanana(g.mouse, 20, g.Texture("banana"), g.Space))
		}
		if g.Keys[glfw.KeyQ] {
			g.Bombs = append(g.Bombs, NewBomb(g.mouse, 20, g.Space))
		}
		if g.Keys[glfw.KeyF] {
			g.fullscreen = !g.fullscreen
			openGlWindow.SetFullscreen(g.fullscreen)
		}
		if g.Keys[glfw.KeyEnter] {
			i := len(g.Players)
			pos := cp.Vector{center.X + rand.Float64()*10, center.Y + rand.Float64()*10}
			g.Players = append(g.Players, NewPlayer(pos, playerRadius, g))
			g.Players[i].Color = eng.NextColor()
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
		// give the mouse click a little radius to make it easier to click small shapes.
		const clickRadius = 5

		if button == glfw.MouseButton1 {
			if action == glfw.Press {
				info := g.Space.PointQueryNearest(g.mouse, clickRadius, NotGrabbableFilter)

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
					leftDown := g.mouse.Clone()
					g.leftDown = &leftDown
					seg := cp.NewSegment(g.Space.StaticBody, *g.leftDown, g.mouse, wallWidth)
					seg.SetElasticity(1)
					seg.SetFriction(wallFriction)
					g.drawingWallShape = seg.Class.(*cp.Segment)
					g.Walls = append(g.Walls, g.drawingWallShape)
				}
				return
			}
			// mouse up
			if g.mouseJoint != nil {
				g.Space.RemoveConstraint(g.mouseJoint)
				g.mouseJoint = nil
				return
			}
			if g.leftDown != nil {
				g.leftDown = nil
			}
			return
		}

		if button == glfw.MouseButton2 {
			if action == glfw.Press {
				rightDown := g.mouse.Clone()
				g.rightDown = &rightDown
			} else {
				g.rightDown = nil

				info := g.Space.PointQueryNearest(g.mouse, clickRadius, NotGrabbableFilter)

				if info.Shape != nil {
					if segment, ok := info.Shape.Class.(*cp.Segment); ok {
						for i, w := range g.Walls {
							if segment == w {
								g.Walls = append(g.Walls[:i], g.Walls[i+1:]...)
								g.Space.AddPostStepCallback(func(space *cp.Space, key interface{}, data interface{}) {
									space.RemoveShape(w.Shape)
									space.RemoveBody(w.Body())
								}, nil, nil)
								break
							}
						}
					}
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

	if g.leftDown != nil {
		g.drawingWallShape.SetEndpoints(*g.leftDown, g.mouse)
	} else if g.drawingWallShape != nil {
		g.Space.AddShape(g.drawingWallShape.Shape)
		g.drawingWallShape = nil
	}

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
		g.TextRenderer.Use().SetMat4("projection", mgl32.Ortho2D(0, float32(g.window.Width), float32(g.window.Height), 0))
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
		g.Bombs[i].Draw(g, alpha)
	}
	for i := range g.Players {
		g.Players[i].Draw(g, alpha)
	}
	g.CPRenderer.ClearRenderer()
	for _, wall := range g.Walls {
		g.CPRenderer.DrawFatSegment(wall.A(), wall.B(), wall.Radius(), eng.DefaultOutline, eng.DefaultFill)
	}
	if len(g.Walls) > 0 {
		g.CPRenderer.FlushRenderer()
	}
	//}

	if len(g.Players) == 0 {
		g.TextRenderer.Print("Connect controllers or press ENTER to use keyboard", float64(g.window.Width)/2.-250., float64(g.window.Height)/2., 1)
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
	g.Space.Iterations = 10
	g.Space.SetGravity(cp.Vector{0, Gravity})

	bananaCollisionHandler := g.Space.NewCollisionHandler(collisionBanana, collisionPlayer)
	bananaCollisionHandler.PreSolveFunc = BananaPreSolve
	bananaCollisionHandler.UserData = g

	bombCollisionHandler := g.Space.NewCollisionHandler(collisionBomb, collisionPlayer)
	bombCollisionHandler.PreSolveFunc = BombPreSolve

	center := cp.Vector{worldWidth / 2, worldHeight / 2}

	// load the initial level
	if err := g.loadLevel("assets/levels/initial.json"); err != nil {
		panic(err)
	}

	for _, p := range g.Players {
		pos := cp.Vector{center.X + rand.Float64()*10, center.Y + rand.Float64()*10}
		p.Reset(pos, playerRadius, g)
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

func (g *Game) saveLevel(filename string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Println(err)
		return
	}
	type entry struct {
		A, B cp.Vector
	}
	var data []entry
	for _, w := range g.Walls {
		data = append(data, entry{w.A(), w.B()})
	}
	if err = json.NewEncoder(file).Encode(data); err != nil {
		log.Println(err)
	}
}

const wallFriction = 100

func (g *Game) loadLevel(name string) error {
	file, err := os.Open(name)
	if err != nil {
		log.Println(err)
		return err
	}
	type entry struct {
		A, B cp.Vector
	}
	var data []entry
	if err = json.NewDecoder(file).Decode(&data); err != nil {
		log.Println(err)
		return err
	}

	g.Walls = []*cp.Segment{}
	for _, w := range data {
		var seg *cp.Shape
		seg = g.Space.AddShape(cp.NewSegment(g.Space.StaticBody, w.A, w.B, wallWidth))
		seg.SetElasticity(1)
		seg.SetFriction(wallFriction)
		g.Walls = append(g.Walls, seg.Class.(*cp.Segment))
	}

	return nil
}
