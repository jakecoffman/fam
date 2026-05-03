package eng

import (
	"fmt"
	"runtime"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

// PhysicsDt is the fixed physics timestep (seconds per sub-step).
const PhysicsDt = 1.0 / 120.0

type Scene interface {
	New(window *OpenGlWindow)
	Render(float64)
	Update(float64)
	Close()
}

func Run(scene Scene) {
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

	window := NewOpenGlWindow()

	defer window.Destroy()

	if err := gl.Init(); err != nil {
		panic(err)
	}
	glfw.SwapInterval(1)

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	const dt = PhysicsDt
	const maxAccumulator = dt * 5

	currentTime := glfw.GetTime()
	accumulator := 0.0

	frames := 0
	lastFps := glfw.GetTime()

	scene.New(window)
	defer scene.Close()

	for !window.ShouldClose() {
		frames++
		glfw.PollEvents()

		newTime := glfw.GetTime()
		if elapsed := newTime - lastFps; elapsed >= 1 {
			fps := float64(frames) / elapsed
			window.SetTitle(fmt.Sprintf("Game | %d FPS", int(fps+0.5)))
			frames = 0
			lastFps = newTime
		}
		frameTime := newTime - currentTime
		if frameTime > .25 {
			frameTime = .25
		}
		currentTime = newTime
		accumulator += frameTime
		// Cap the accumulator to avoid a "spiral of death" after stalls.
		if accumulator > maxAccumulator {
			accumulator = maxAccumulator
		}

		for accumulator >= dt {
			scene.Update(dt)
			accumulator -= dt
		}

		gl.ClearColor(0, 0, 0, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		alpha := accumulator / dt
		scene.Render(alpha)
		window.SwapBuffers()
	}
}
