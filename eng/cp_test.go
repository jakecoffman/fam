package eng

import (
	"testing"

	"github.com/jakecoffman/cp/v2"
)

func TestCPRendererInvalidatesMesh(t *testing.T) {
	renderer := &CPRenderer{}
	draws := []struct {
		name string
		draw func()
	}{
		{"circle", func() { renderer.DrawCircle(cp.Vector{}, 0, 10, DefaultOutline, DefaultFill) }},
		{"segment", func() { renderer.DrawFatSegment(cp.Vector{}, cp.Vector{X: 20}, 2, DefaultOutline, DefaultFill) }},
		{"polygon", func() {
			renderer.DrawPolygon(3, []cp.Vector{{}, {X: 20}, {Y: 20}}, 0, DefaultOutline, DefaultFill)
		}},
		{"dot", func() { renderer.DrawDot(10, cp.Vector{}, DefaultFill) }},
	}
	for _, test := range draws {
		t.Run(test.name, func(t *testing.T) {
			renderer.dirty = false
			test.draw()
			if !renderer.dirty || len(renderer.triangles) == 0 {
				t.Error("drawing did not invalidate mesh")
			}
			renderer.dirty = false
			renderer.Clear()
			if !renderer.dirty || len(renderer.triangles) != 0 {
				t.Error("clearing did not invalidate and empty mesh")
			}
			renderer.Flush()
		})
	}
}
