package fam

import (
	"fmt"
	"os"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/inkyblackness/imgui-go"
	"github.com/jakecoffman/fam/platforms"
	"github.com/jakecoffman/fam/renderers"
)

type Gui struct {
	*imgui.Context
	platform       *platforms.GLFW
	renderer       *renderers.OpenGL3

	game *Game

	showDemoWindow bool
	showAnotherWindow bool
}

func NewGui(game *Game) *Gui {
	g := &Gui{
		Context: imgui.CreateContext(nil),
		game: game,
	}
	io := imgui.CurrentIO()

	platform, err := platforms.NewGLFW(io, game.window.Window)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
	g.platform = platform

	renderer, err := renderers.NewOpenGL3(io)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
	g.renderer = renderer

	imgui.CurrentIO().SetClipboard(clipboard{platform: g.platform})

	g.showDemoWindow = false
	g.showAnotherWindow = false

	return g
}

func (gui *Gui) Destroy() {
	gui.renderer.Dispose()
	gui.Context.Destroy()
}

// Platform covers mouse/keyboard/gamepad inputs, cursor shape, timing, windowing.
type Platform interface {
	// ShouldStop is regularly called as the abort condition for the program loop.
	ShouldStop() bool
	// ProcessEvents is called once per render loop to dispatch any pending events.
	ProcessEvents()
	// DisplaySize returns the dimension of the display.
	DisplaySize() [2]float32
	// FramebufferSize returns the dimension of the framebuffer.
	FramebufferSize() [2]float32
	// NewFrame marks the begin of a render pass. It must update the imgui IO state according to user input (mouse, keyboard, ...)
	NewFrame()
	// PostRender marks the completion of one render pass. Typically this causes the display buffer to be swapped.
	PostRender()
	// ClipboardText returns the current text of the clipboard, if available.
	ClipboardText() (string, error)
	// SetClipboardText sets the text as the current text of the clipboard.
	SetClipboardText(text string)
}

type clipboard struct {
	platform Platform
}

func (board clipboard) Text() (string, error) {
	return board.platform.ClipboardText()
}

func (board clipboard) SetText(text string) {
	board.platform.SetClipboardText(text)
}

// Renderer covers rendering imgui draw data.
type Renderer interface {
	// PreRender causes the display buffer to be prepared for new output.
	PreRender(clearColor [4]float32)
	// Render draws the provided imgui draw data.
	Render(displaySize [2]float32, framebufferSize [2]float32, drawData imgui.DrawData)
}

// Run implements the main program loop of the demo. It returns when the platform signals to stop.
// This demo application shows some basic features of ImGui, as well as exposing the standard demo window.
func (gui *Gui) Render() {
	p := gui.platform

	p.NewFrame()
	imgui.NewFrame()

	// 1. Show the big demo window (Most of the sample code is in ImGui::ShowDemoWindow()!
	// You can browse its code to learn more about Dear ImGui!).
	if gui.showDemoWindow {
		imgui.ShowDemoWindow(&gui.showDemoWindow)
	}

	// 2. Show a simple window that we create ourselves. We use a Begin/End pair to created a named window.
	{
		imgui.Begin("Menu")

		if imgui.Button("Resume game") {
			gui.game.unpause()
		}

		// LMB, RMB action
		items := []string{actionBanana, actionBomb}

		if imgui.BeginComboV("LMB", gui.game.lmbAction, imgui.ComboFlagNoArrowButton) {
			for _, item := range items {
				isSelected := gui.game.lmbAction == item
				if imgui.SelectableV(item, isSelected, 0, imgui.Vec2{}) {
					gui.game.lmbAction = item
				}
				if isSelected {
					imgui.SetItemDefaultFocus()
				}
			}
			imgui.EndCombo()
		}
		if imgui.BeginComboV("RMB", gui.game.rmbAction, imgui.ComboFlagNoArrowButton) {
			for _, item := range items {
				isSelected := gui.game.rmbAction == item
				if imgui.SelectableV(item, isSelected, 0, imgui.Vec2{}) {
					gui.game.rmbAction = item
				}
				if isSelected {
					imgui.SetItemDefaultFocus()
				}
			}
			imgui.EndCombo()
		}

		if imgui.Button("Reset objects") {
			gui.game.reset()
		}

		imgui.Checkbox("Render Physics", &gui.game.shouldRenderCp)
		if imgui.Checkbox("Vsync", &gui.game.vsync) {
			if gui.game.vsync {
				glfw.SwapInterval(1)
			} else {
				glfw.SwapInterval(0)
			}
		}

		if imgui.ButtonV("Quit", imgui.Vec2{200, 20}) {
			//gui.counter++
			gui.game.window.SetShouldClose(true)
		}
		//imgui.SameLine()
		//imgui.Text(fmt.Sprintf("counter = %d", gui.counter))

		// TODO add text of FPS based on IO.Framerate()

		imgui.End()
	}

	// 3. Show another simple window.
	if gui.showAnotherWindow {
		// Pass a pointer to our bool variable (the window will have a closing button that will clear the bool when clicked)
		imgui.BeginV("Another window", &gui.showAnotherWindow, 0)

		imgui.Text("Hello from another window!")
		if imgui.Button("Close Me") {
			gui.showAnotherWindow = false
		}
		imgui.End()
	}

	imgui.Render()
	gui.renderer.Render(p.DisplaySize(), p.FramebufferSize(), imgui.RenderedDrawData())
}
