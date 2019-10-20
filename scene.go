package fam

import (
	"fmt"
	"runtime"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

type Scene interface {
	New(width, height int, window *glfw.Window)
	Render(float64)
	Update(float64)
	Close()
}

func Run(scene Scene, width, height int) {
	runtime.LockOSThread()

	// glfw: initialize and configure
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	defer glfw.Terminate()
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)

	if runtime.GOOS == "darwin" {
		glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	}

	// glfw window creation
	monitor := glfw.GetPrimaryMonitor()
	videoMode := monitor.GetVideoMode()
	window, err := glfw.CreateWindow(videoMode.Width, videoMode.Height, "Game", monitor, nil)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()
	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		panic(err)
	}
	glfw.SwapInterval(1)

	gl.Enable(gl.CULL_FACE)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	const dt = 1. / 60.
	currentTime := glfw.GetTime()
	accumulator := 0.0

	frames := 0
	var lastFps float64

	scene.New(videoMode.Width, videoMode.Height, window)

	for !window.ShouldClose() {
		frames++
		glfw.PollEvents()

		newTime := glfw.GetTime()
		if newTime-lastFps > 1 {
			window.SetTitle(fmt.Sprintf("Game | %d FPS", frames))
			frames = 0
			lastFps = newTime
		}
		frameTime := newTime - currentTime
		if frameTime > .25 {
			frameTime = .25
		}
		currentTime = newTime
		accumulator += frameTime

		for accumulator >= dt {
			scene.Update(dt)
			accumulator -= dt
		}

		gl.ClearColor(0, 0, 0, 0.5)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		alpha := accumulator / dt
		scene.Render(alpha)
		window.SwapBuffers()
	}

	scene.Close()
}
