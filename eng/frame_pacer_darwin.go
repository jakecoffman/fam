package eng

/*
#cgo LDFLAGS: -framework Cocoa -framework CoreVideo
#include <stdint.h>

typedef struct FamFrameClock FamFrameClock;
FamFrameClock *famFrameClockCreate(uintptr_t window, int *status);
int famFrameClockWait(FamFrameClock *clock);
int famFrameClockDestroy(FamFrameClock *clock);
*/
import "C"

import (
	"fmt"

	"github.com/go-gl/glfw/v3.2/glfw"
)

type framePacer struct {
	clock *C.FamFrameClock
}

func (p *framePacer) setEnabled(window *glfw.Window, enabled bool) {
	if enabled && p.clock == nil {
		var status C.int
		p.clock = C.famFrameClockCreate(C.uintptr_t(window.GetCocoaWindow()), &status)
		if p.clock == nil {
			panic(fmt.Sprintf("creating macOS display clock: %d", int(status)))
		}
	} else if !enabled {
		p.close()
	}
	// NSOpenGL swap intervals can deliver alternating short/long frames on macOS.
	// The display clock paces rendering instead of the driver's swap interval.
	glfw.SwapInterval(0)
}

func (p *framePacer) wait() bool {
	if p.clock == nil {
		return true
	}
	status := C.famFrameClockWait(p.clock)
	if status < 0 {
		panic(fmt.Sprintf("waiting for macOS display clock: %d", int(status)))
	}
	return status != 0
}

func (p *framePacer) close() {
	if p.clock != nil {
		if status := C.famFrameClockDestroy(p.clock); status != 0 {
			panic(fmt.Sprintf("stopping macOS display clock: %d", int(status)))
		}
		p.clock = nil
	}
}
