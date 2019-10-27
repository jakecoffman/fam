package eng

import "github.com/go-gl/glfw/v3.2/glfw"

type OpenGlWindow struct {
	*glfw.Window
	Monitor *glfw.Monitor

	X, Y                          int
	Width, Height                 int
	ViewportWidth, ViewPortHeight int

	UpdateViewport bool
}

const (
	initialWindowWidth  = 1024
	initialWindowHeight = 576
)

func NewOpenGlWindow() *OpenGlWindow {
	window := &OpenGlWindow{}
	var err error
	window.Window, err = glfw.CreateWindow(initialWindowWidth, initialWindowHeight, "Game", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()
	window.SetSizeCallback(func(w *glfw.Window, width, height int) {
		window.UpdateViewport = true
	})

	window.Monitor = glfw.GetPrimaryMonitor()
	window.Width, window.Height = window.GetSize()
	window.X, window.Y = window.GetPos()
	window.UpdateViewport = true
	return window
}

func (w *OpenGlWindow) IsFullscreen() bool {
	return w.GetMonitor() != nil
}

func (w *OpenGlWindow) SetFullscreen(fullscreen bool) {
	if fullscreen == w.IsFullscreen() {
		return
	}

	if fullscreen {
		w.X, w.Y = w.GetPos()
		w.Width, w.Height = w.GetSize()

		mode := w.Monitor.GetVideoMode()
		w.SetMonitor(w.Monitor, 0, 0, mode.Width, mode.Height, 0)
	} else {
		w.SetMonitor(nil, w.X, w.Y, w.Width, w.Height, 0)
	}

	w.UpdateViewport = true
}

func (w *OpenGlWindow) Resize() {
	w.UpdateViewport = true
}
