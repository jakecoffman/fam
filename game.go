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
	"time"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp/v2"
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
	collisionWall
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Game struct {
	state      int
	Keys       map[glfw.Key]bool
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

	drawingWallShape *Wall

	Space *cp.Space

	Players []*Player
	Bananas []*Banana
	Bombs   []*Bomb
	Walls   []*Wall

	*eng.ResourceManager

	ParticleGenerator *eng.ParticleGenerator
	SpriteRenderer    *eng.SpriteRenderer
	CPRenderer        *eng.CPRenderer
	WallRenderer      *eng.CPRenderer
	TextRenderer      *eng.TextRenderer

	shouldRenderCp bool
	wallsDirty     bool

	chaseBananaMode bool
	randomBombMode  bool

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
	playerRadius = 25.0
)

func (g *Game) New(openGlWindow *eng.OpenGlWindow) {
	g.vsync = true
	g.window = openGlWindow
	g.Keys = make(map[glfw.Key]bool)
	g.mouseBody = cp.NewKinematicBody()
	openGlWindow.SetVsync(g.vsync)

	g.ResourceManager = eng.NewResourceManager()

	g.LoadShader("assets/shaders/main.vs.glsl", "assets/shaders/main.fs.glsl", "sprite")
	g.LoadShader("assets/shaders/particle.vs.glsl", "assets/shaders/particle.fs.glsl", "particle")
	g.LoadShader("assets/shaders/cp.vs.glsl", "assets/shaders/cp.fs.glsl", "cp")
	g.LoadShader("assets/shaders/text.vs.glsl", "assets/shaders/text.fs.glsl", "text")

	g.projection = mgl32.Ortho(0, worldWidth, worldHeight, 0, -1, 1)
	g.Shader("sprite").Use().SetInt("image", 0).SetMat4("projection", g.projection)
	g.Shader("particle").Use().SetInt("sprite", 0).SetMat4("projection", g.projection)
	g.CPRenderer = eng.NewCPRenderer(g.Shader("cp"), g.projection)
	g.WallRenderer = eng.NewCPRenderer(g.Shader("cp"), g.projection)
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
			g.connectJoystick(glfw.Joystick(joy))
		} else {
			log.Println("Joystick disconnected", joy)
		}
	})

	g.discoverJoysticks(glfw.JoystickPresent)

	g.state = stateActive

	openGlWindow.SetCursorPosCallback(func(w *glfw.Window, xpos float64, ypos float64) {
		ww, wh := w.GetSize()
		g.mouse = g.MouseToSpace(xpos, ypos, ww, wh)
	})

	openGlWindow.SetKeyCallback(func(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		g.handleKey(key, action)
	})

	openGlWindow.SetMouseButtonCallback(func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
		g.handleMouseButton(button, action)
	})

	// Install ImGui last so it can chain the game's input callbacks.
	g.gui = NewGui(g)
}

func (g *Game) addPlayer(joy glfw.Joystick) *Player {
	pos := cp.Vector{worldWidth/2 + rand.Float64()*10, worldHeight/2 + rand.Float64()*10}
	player := NewPlayer(pos, playerRadius, g)
	player.Color = eng.NextColor()
	player.Joystick = joy
	g.Players = append(g.Players, player)
	return player
}

func (g *Game) connectJoystick(joy glfw.Joystick) {
	for _, player := range g.Players {
		if player.Joystick == joy {
			log.Println("Joystick reconnected", joy)
			return
		}
	}
	log.Println("Joystick connected", joy)
	g.addPlayer(joy)
}

func (g *Game) discoverJoysticks(present func(glfw.Joystick) bool) {
	for i := 0; i < 16; i++ {
		joy := glfw.Joystick(i)
		if present(joy) {
			g.connectJoystick(joy)
		}
	}
}

func (g *Game) handleKey(key glfw.Key, action glfw.Action) {
	if action == glfw.Press {
		g.Keys[key] = true
	} else if action == glfw.Release {
		delete(g.Keys, key)
	}
	if action != glfw.Press {
		return
	}
	if key == glfw.KeyEscape {
		if g.state == stateActive {
			g.pause()
		} else {
			g.unpause()
		}
		return
	}
	if g.state != stateActive {
		return
	}
	switch key {
	case glfw.KeyE:
		g.Bananas = append(g.Bananas, NewBanana(g, g.mouse, 20))
	case glfw.KeyQ:
		g.Bombs = append(g.Bombs, NewBomb(g.mouse, 20, g.Space))
	case glfw.KeyF:
		g.fullscreen = !g.fullscreen
		g.window.SetFullscreen(g.fullscreen)
	case glfw.KeyEnter:
		g.addPlayer(glfw.Joystick(-1))
	}
}

func (g *Game) handleMouseButton(button glfw.MouseButton, action glfw.Action) {
	if g.state != stateActive {
		return
	}
	const clickRadius = 5
	if button == glfw.MouseButton1 {
		if action == glfw.Press {
			info := g.Space.PointQueryNearest(g.mouse, clickRadius, NotGrabbableFilter)
			if info.Shape != nil && info.Shape.Body().Mass() < cp.INFINITY {
				nearest := g.mouse
				if info.Distance > 0 {
					nearest = info.Point
				}
				body := info.Shape.Body()
				g.mouseJoint = cp.NewPivotJoint2(g.mouseBody, body, cp.Vector{}, body.WorldToLocal(nearest))
				g.mouseJoint.SetMaxForce(50000)
				g.mouseJoint.SetErrorBias(math.Pow(1.0-0.15, 1.0/eng.PhysicsDt))
				g.Space.AddConstraint(g.mouseJoint)
			} else {
				leftDown := g.mouse
				g.leftDown = &leftDown
				g.drawingWallShape = NewWall(g, *g.leftDown, g.mouse)
			}
		} else if action == glfw.Release {
			g.finishMouseDrag()
		}
		return
	}
	if button == glfw.MouseButton2 {
		if action == glfw.Press {
			rightDown := g.mouse
			g.rightDown = &rightDown
		} else if action == glfw.Release && g.rightDown != nil {
			g.rightDown = nil
			info := g.Space.PointQueryNearest(g.mouse, clickRadius, NotGrabbableFilter)
			for i, wall := range g.Walls {
				if wall.Shape == info.Shape {
					g.removeWall(i)
					break
				}
			}
		}
	}
}

func (g *Game) finishMouseDrag() {
	if g.drawingWallShape != nil && !g.leftDown.Equal(g.mouse) {
		wall := g.drawingWallShape
		wall.SetEndpoints(*g.leftDown, g.mouse)
		g.Space.AddShape(wall.Shape)
		g.Walls = append(g.Walls, wall)
		g.wallsDirty = true
	}
	g.cancelMouseDrag()
}

func (g *Game) cancelMouseDrag() {
	if g.mouseJoint != nil {
		g.Space.RemoveConstraint(g.mouseJoint)
		g.mouseJoint = nil
	}
	g.leftDown = nil
	g.rightDown = nil
	g.drawingWallShape = nil
}

func (g *Game) removeWall(index int) {
	g.Space.RemoveShape(g.Walls[index].Shape)
	copy(g.Walls[index:], g.Walls[index+1:])
	g.Walls[len(g.Walls)-1] = nil
	g.Walls = g.Walls[:len(g.Walls)-1]
	g.wallsDirty = true
}

func (g *Game) Update(dt float64) {
	if g.state == statePause {
		return
	}

	if g.chaseBananaMode && len(g.Bananas) == 0 {
		x := rand.Intn(worldWidth)
		y := rand.Intn(worldHeight)
		banana := NewBanana(g, cp.Vector{float64(x), float64(y)}, 20)
		banana.SetVelocity(float64(rand.Intn(2000)-1000), float64(rand.Intn(2000)-1000))
		g.Bananas = append(g.Bananas, banana)
	}
	if g.randomBombMode && len(g.Bombs) == 0 {
		x := rand.Intn(worldWidth)
		y := rand.Intn(worldHeight)
		bomb := NewBomb(cp.Vector{float64(x), float64(y)}, 20, g.Space)
		bomb.SetVelocity(float64(rand.Intn(2000)-1000), float64(rand.Intn(2000)-1000))
		g.Bombs = append(g.Bombs, bomb)
	}

	// update mouse body
	newPoint := g.mouseBody.Position().Lerp(g.mouse, 0.25)
	g.mouseBody.SetVelocityVector(newPoint.Sub(g.mouseBody.Position()).Mult(1.0 / eng.PhysicsDt))
	g.mouseBody.SetPosition(newPoint)

	if g.leftDown != nil {
		g.drawingWallShape.SetEndpoints(*g.leftDown, g.mouse)
	}

	for i := range g.Bombs {
		g.Bombs[i].Update(g, dt)
	}
	// Filter out gone bombs without discarding live ones.
	out := g.Bombs[:0]
	for _, b := range g.Bombs {
		if b.state != bombStateGone {
			out = append(out, b)
		}
	}
	g.Bombs = out
	for i := range g.Bananas {
		g.Bananas[i].Update(g, dt)
	}
	for i := range g.Players {
		g.Players[i].Update(g, dt)
	}

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

	g.SpriteRenderer.DrawSprite(g.Texture("background"), mgl32.Vec2{worldWidth / 2, worldHeight / 2}, mgl32.Vec2{worldWidth, worldHeight}, 0, eng.White)
	g.SpriteRenderer.Flush()

	g.CPRenderer.Clear()
	if g.shouldRenderCp {
		g.CPRenderer.DrawSpace(g.Space)
	} else {
		g.updateWallMesh()
		g.WallRenderer.Flush()
	}
	if g.drawingWallShape != nil {
		g.drawingWallShape.Draw(g.CPRenderer)
	}
	g.CPRenderer.Flush()

	if len(g.Players) == 0 {
		g.TextRenderer.Print("Connect controllers or press ENTER to use keyboard", float64(g.window.Width)/2.-250., float64(g.window.Height)/2., 1)
	}

	//g.SpriteRenderer.DrawSprite(g.Texture("banana"), V(g.mouse), mgl32.Vec2{100, 100}, 0, mgl32.Vec3{1, 0, 0})

	for i := range g.Bananas {
		g.Bananas[i].Draw(g.SpriteRenderer, alpha)
	}
	for i := range g.Bombs {
		g.Bombs[i].Draw(g, alpha)
	}
	for i := range g.Players {
		g.Players[i].Draw(g, alpha)
	}
	g.SpriteRenderer.Flush()

	g.gui.Render()
}

func (g *Game) updateWallMesh() {
	if !g.wallsDirty {
		return
	}
	g.WallRenderer.Clear()
	for _, wall := range g.Walls {
		wall.Draw(g.WallRenderer)
	}
	g.wallsDirty = false
}

func (g *Game) Close() {
	g.gui.Destroy()
	g.SpriteRenderer.Destroy()
	g.CPRenderer.Destroy()
	g.WallRenderer.Destroy()
	g.Clear()
}

func (g *Game) pause() {
	g.finishMouseDrag()
	g.state = statePause
}

func (g *Game) unpause() {
	g.state = stateActive
}

func (g *Game) reset() {
	g.cancelMouseDrag()
	g.Walls = nil
	g.Space = cp.NewSpace()
	g.Space.Iterations = 10
	g.Space.SetGravity(cp.Vector{0, Gravity})

	bananaCollisionHandler := g.Space.NewCollisionHandler(collisionBanana, collisionPlayer)
	bananaCollisionHandler.PreSolveFunc = BananaPreSolve
	bananaCollisionHandler.UserData = g

	bombCollisionHandler := g.Space.NewWildcardCollisionHandler(collisionBomb)
	bombCollisionHandler.PreSolveFunc = BombPreSolve

	g.Space.NewWildcardCollisionHandler(collisionWall).PreSolveFunc = WallPreSolve

	center := cp.Vector{worldWidth / 2, worldHeight / 2}

	// load the initial level
	if err := g.loadLevel("assets/levels/initial.json"); err != nil {
		panic(err)
	}

	var players []*Player
	for _, p := range g.Players {
		if p.Joystick == glfw.Joystick(-1) {
			// remove players created with "enter" for when the kids make too many players
			continue
		}
		pos := cp.Vector{center.X + rand.Float64()*10, center.Y + rand.Float64()*10}
		p.Reset(pos, playerRadius, g)
		players = append(players, p)
	}
	g.Players = players
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
	defer func() {
		if err := file.Close(); err != nil {
			log.Println(err)
		}
	}()
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

func (g *Game) loadLevel(name string) error {
	file, err := os.Open(name)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Println(err)
		}
	}()
	type entry struct {
		A, B cp.Vector
	}
	var data []entry
	if err = json.NewDecoder(file).Decode(&data); err != nil {
		log.Println(err)
		return err
	}

	g.cancelMouseDrag()
	for _, wall := range g.Walls {
		g.Space.RemoveShape(wall.Shape)
	}
	g.Walls = nil
	for _, w := range data {
		wall := NewWall(g, w.A, w.B)
		g.Space.AddShape(wall.Segment.Shape)
		g.Walls = append(g.Walls, wall)
	}
	g.wallsDirty = true

	return nil
}
