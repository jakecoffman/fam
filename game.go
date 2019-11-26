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
	"github.com/inkyblackness/imgui-go"
	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/eng"
)

var GrabbableMaskBit uint = 1 << 31

//var GrabFilter = cp.ShapeFilter{
//	cp.NO_GROUP, GrabbableMaskBit, GrabbableMaskBit,
//}

var NotGrabbableFilter = cp.ShapeFilter{
	cp.NO_GROUP, ^GrabbableMaskBit, ^GrabbableMaskBit,
}

var PlayerMaskBit uint = 1 << 30

var PlayerFilter = cp.ShapeFilter{
	cp.NO_GROUP, PlayerMaskBit, PlayerMaskBit,
}

//var NotPlayerFilter = cp.ShapeFilter{
//	cp.NO_GROUP, ^PlayerMaskBit, ^PlayerMaskBit,
//}

const (
	_ = iota
	collisionPlayer
	collisionBanana
	collisionBomb
	collisionWall
)

type Game struct {
	state      int
	Keys       [1024]bool
	vsync      bool
	fullscreen bool
	window     *eng.OpenGlWindow
	gui        *Gui

	projection mgl32.Mat4

	mouse Mouse

	Space *cp.Space

	Objects *eng.ObjectSystem
	Players *PlayerSystem
	Bananas *BananaSystem
	Bombs   *BombSystem
	Walls   *WallSystem

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

const (
	gameStateActive = iota
	gameStatePaused
)

var centerOfWorld = cp.Vector{worldWidth/2, worldHeight/2}

func (g *Game) New(openGlWindow *eng.OpenGlWindow) {
	g.vsync = true
	g.window = openGlWindow
	g.gui = NewGui(g)
	g.Keys = [1024]bool{}
	g.mouse.New()

	g.ResourceManager = eng.NewResourceManager()

	g.LoadShader("assets/shaders/main.vs.glsl", "assets/shaders/main.fs.glsl", "sprite")
	g.LoadShader("assets/shaders/particle.vs.glsl", "assets/shaders/particle.fs.glsl", "particle")
	g.LoadShader("assets/shaders/cp.vs.glsl", "assets/shaders/cp.fs.glsl", "cp")
	g.LoadShader("assets/shaders/text.vs.glsl", "assets/shaders/text.fs.glsl", "text")

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

	glfw.SetJoystickCallback(g.joystickCallback)

	// run though the connected joysticks and add players for each one
	for i := 0; i < 16; i++ {
		joy := glfw.Joystick(i)
		if !glfw.JoystickPresent(joy) {
			break
		}
		g.Players.Add(centerOfWorld, eng.NextColor(), joy)
	}

	g.state = gameStateActive

	openGlWindow.SetCursorPosCallback(g.mouseMoveCallback)
	openGlWindow.SetKeyCallback(g.keyCallback)
	openGlWindow.SetMouseButtonCallback(g.mouseCallback)
}

func (g *Game) Update(dt float64) {
	g.mouse.Update(dt)
	g.Walls.Update(dt)
	if g.state == gameStatePaused {
		return
	}
	g.Bombs.Update(dt)
	//g.Bananas.Update(dt)
	g.Players.Update(dt)
	g.Objects.Update(dt, worldWidth, worldHeight)
	g.Space.Step(dt)
}

func (g *Game) Render(alpha float64) {
	if g.window.UpdateViewport {
		g.window.UpdateViewport = false
		g.window.ViewportWidth, g.window.ViewPortHeight = g.window.GetFramebufferSize()
		gl.Viewport(0, 0, int32(g.window.ViewportWidth), int32(g.window.ViewPortHeight))
		g.TextRenderer.Use().SetMat4("projection", mgl32.Ortho2D(0, float32(g.window.Width), float32(g.window.Height), 0))
		log.Printf("update viewport %#v\n", g.window)
	}

	g.SpriteRenderer.DrawSprite(g.Texture("background"), mgl32.Vec2{worldWidth/2, worldHeight/2}, mgl32.Vec2{worldWidth, worldHeight}, 0, eng.White)

	{
		g.CPRenderer.Clear()
		if g.shouldRenderCp {
			g.CPRenderer.DrawSpace(g.Space)
		}
		g.Walls.Draw(alpha)
		g.CPRenderer.Flush()
	}

	if g.state == gameStatePaused {
		g.TextRenderer.Print("Paused (press P to unpause)", float64(g.window.Width)/2.-150., float64(g.window.Height)/2., 1)
	} else if len(g.Players.players) == 0 {
		g.TextRenderer.Print("Connect controllers or press ENTER to use keyboard", float64(g.window.Width)/2.-250., float64(g.window.Height)/2., 1)
	}

	//g.SpriteRenderer.DrawSprite(g.Texture("banana"), V(g.worldPos), mgl32.Vec2{100, 100}, 0, mgl32.Vec3{1, 0, 0})

	g.Bananas.Draw(alpha)
	g.Bombs.Draw(alpha)
	g.Players.Draw(alpha)

	g.gui.Render()
}

func (g *Game) Close() {
	g.gui.Destroy()
	g.Clear()
}

func (g *Game) pause() {
	g.state = gameStatePaused
}

func (g *Game) unpause() {
	g.state = gameStateActive
}

func (g *Game) reset() {
	g.Space = cp.NewSpace()
	g.Space.Iterations = 10
	g.Space.SetGravity(cp.Vector{0, Gravity})

	if g.Objects != nil {
		g.Objects.Reset(g.Space)
	} else {
		g.Objects = eng.NewObjectSystem(g.Space)
	}
	if g.Players != nil {
		g.Players.Reset()
	} else {
		g.Players = NewPlayerSystem(g)
	}

	g.Bananas = NewBananaSystem(g)
	g.Bombs = NewBombSystem(g)
	g.Walls = NewWallSystem(g)

	if err := g.loadLevel("assets/levels/initial.json"); err != nil {
		panic(err)
	}
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
	for _, w := range g.Walls.walls {
		data = append(data, entry{w.A(), w.B()})
	}
	if err = json.NewEncoder(file).Encode(data); err != nil {
		log.Println(err)
	}
}

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

	for _, w := range data {
		wall := g.Walls.Add(w.A, w.B)
		g.Space.AddShape(wall.Shape)
	}

	return nil
}

func (g *Game) mouseCallback(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
	if imgui.CurrentIO().WantCaptureMouse() {
		// the gui is handling the worldPos action
		return
	}
	if g.state != gameStateActive {
		return
	}
	// give the worldPos click a little radius to make it easier to click small shapes.
	const clickRadius = 5

	if button == glfw.MouseButton1 {
		if action == glfw.Press {
			info := g.Space.PointQueryNearest(g.mouse.worldPos, clickRadius, NotGrabbableFilter)

			if info.Shape != nil && info.Shape.Body().Mass() < cp.INFINITY {
				var nearest cp.Vector
				if info.Distance > 0 {
					nearest = info.Point
				} else {
					nearest = g.mouse.worldPos
				}

				body := info.Shape.Body()
				g.mouse.joint = cp.NewPivotJoint2(g.mouse.body, body, cp.Vector{}, body.WorldToLocal(nearest))
				g.mouse.joint.SetMaxForce(50000)
				g.mouse.joint.SetErrorBias(math.Pow(1.0-0.15, 60.0))
				g.Space.AddConstraint(g.mouse.joint)
			} else {
				leftDown := g.mouse.worldPos.Clone()
				g.mouse.leftDownPos = &leftDown
				g.Walls.drawingWallShape = g.Walls.Add(*g.mouse.leftDownPos, g.mouse.worldPos)
			}
			return
		}
		// worldPos up
		if g.mouse.joint != nil {
			g.Space.RemoveConstraint(g.mouse.joint)
			g.mouse.joint = nil
			return
		}
		if g.mouse.leftDownPos != nil {
			g.mouse.leftDownPos = nil
		}
		return
	}

	if button == glfw.MouseButton2 {
		if action == glfw.Press {
			rightDown := g.mouse.worldPos.Clone()
			g.mouse.rightDownPos = &rightDown
		} else {
			g.mouse.rightDownPos = nil

			info := g.Space.PointQueryNearest(g.mouse.worldPos, clickRadius, NotGrabbableFilter)

			if info.Shape != nil {
				if id, ok := info.Shape.UserData.(eng.EntityID); ok {
					g.Space.AddPostStepCallback(func(space *cp.Space, key interface{}, data interface{}) {
						g.Walls.Remove(id)
					}, nil, nil)
				}
			}
		}
	}
}

func (g *Game) keyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if imgui.CurrentIO().WantCaptureKeyboard() {
		// the gui is handling the keyboard callback
		return
	}
	if key == glfw.KeyEscape && action == glfw.Press {
		g.gui.showMainMenu = !g.gui.showMainMenu
	}
	if key == glfw.KeyP && action == glfw.Press {
		if g.state == gameStateActive {
			g.pause()
		} else {
			g.unpause()
		}
	}
	if g.Keys[glfw.KeyE] {
		g.Bananas.Add().SetPosition(g.mouse.worldPos)
	}
	if g.Keys[glfw.KeyQ] {
		g.Bombs.Add().SetPosition(g.mouse.worldPos)
	}
	if g.Keys[glfw.KeyF] {
		g.fullscreen = !g.fullscreen
		g.window.SetFullscreen(g.fullscreen)
	}
	if g.Keys[glfw.KeyEnter] {
		pos := cp.Vector{centerOfWorld.X + rand.Float64()*10, centerOfWorld.Y + rand.Float64()*10}
		g.Players.Add(pos, eng.NextColor(), glfw.Joystick(-1))
	}
	// store for continuous application
	if key >= 0 && key < 1024 {
		if action == glfw.Press {
			g.Keys[key] = true
		} else if action == glfw.Release {
			g.Keys[key] = false
		}
	}
}

func (g *Game) joystickCallback(joy, event int) {
	if glfw.MonitorEvent(event) == glfw.Connected {
		if joy+1 <= len(g.Players.players) {
			log.Println("Joystick reconnected", joy)
			return
		}
		log.Println("Joystick connected", joy)
		pos := cp.Vector{centerOfWorld.X + rand.Float64()*10, centerOfWorld.Y + rand.Float64()*10}
		g.Players.Add(pos, eng.NextColor(), glfw.Joystick(joy))
	} else {
		log.Println("Joystick disconnected", joy)
	}
}

func (g *Game) mouseMoveCallback(w *glfw.Window, xpos float64, ypos float64) {
	ww, wh := w.GetSize()
	g.mouse.worldPos = g.MouseToSpace(xpos, ypos, ww, wh)
}
