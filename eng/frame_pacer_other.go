//go:build !darwin

package eng

import "github.com/go-gl/glfw/v3.2/glfw"

type framePacer struct{}

func (p *framePacer) setEnabled(window *glfw.Window, enabled bool) {
	if enabled {
		glfw.SwapInterval(1)
	} else {
		glfw.SwapInterval(0)
	}
}

func (p *framePacer) wait() bool { return true }

func (p *framePacer) close() {}
