package eng

import (
	"math"
	"testing"

	"github.com/jakecoffman/cp/v2"
)

func TestObjectInterpolationAcrossFrameRates(t *testing.T) {
	for _, test := range []struct {
		name   string
		frames []float64
	}{
		{"60Hz", []float64{1.0 / 60}},
		{"120Hz", []float64{1.0 / 120}},
		{"144Hz", []float64{1.0 / 144}},
		{"uneven", []float64{0.014, 0.002, 0.009}},
	} {
		t.Run(test.name, func(t *testing.T) {
			space := cp.NewSpace()
			body := space.AddBody(cp.NewBody(1, 1))
			shape := space.AddShape(cp.NewCircle(body, 5, cp.Vector{}))
			body.SetVelocity(200, 0)
			body.SetAngularVelocity(4)
			obj := &Object{Body: body, Shape: shape}
			elapsed, accumulator := 0.0, 0.0
			for frame := range 240 {
				delta := test.frames[frame%len(test.frames)]
				elapsed += delta
				accumulator += delta
				for accumulator >= PhysicsDt {
					obj.Update(space, PhysicsDt, 10000, 10000)
					space.Step(PhysicsDt)
					accumulator -= PhysicsDt
				}
				alpha := accumulator / PhysicsDt
				renderTime := math.Max(0, elapsed-PhysicsDt)
				if got, want := float64(obj.SmoothPos(alpha).X()), 200*renderTime; math.Abs(got-want) > 0.0001 {
					t.Fatalf("frame %d position = %v, want %v", frame, got, want)
				}
				if got, want := obj.SmoothAngle(alpha), 4*renderTime; math.Abs(got-want) > 1e-9 {
					t.Fatalf("frame %d angle = %v, want %v", frame, got, want)
				}
			}
		})
	}
}

func TestObjectUpdateWrap(t *testing.T) {
	tests := []struct {
		name string
		pos  cp.Vector
		want cp.Vector
	}{
		{"left", cp.Vector{X: -20, Y: 50}, cp.Vector{X: 180, Y: 50}},
		{"right", cp.Vector{X: 220, Y: 50}, cp.Vector{X: 20, Y: 50}},
		{"top", cp.Vector{X: 100, Y: -20}, cp.Vector{X: 100, Y: 80}},
		{"bottom", cp.Vector{X: 100, Y: 120}, cp.Vector{X: 100, Y: 20}},
		{"inside", cp.Vector{X: 100, Y: 50}, cp.Vector{X: 100, Y: 50}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			space := cp.NewSpace()
			body := space.AddBody(cp.NewBody(1, 1))
			body.SetPosition(tt.pos)
			shape := space.AddShape(cp.NewCircle(body, 5, cp.Vector{}))
			offset := cp.Vector{X: 12}
			extra := space.AddShape(cp.NewCircle(body, 2, offset))
			obj := &Object{Body: body, Shape: shape}

			obj.Update(space, PhysicsDt, 200, 100)

			if got := body.Position(); !got.Equal(tt.want) {
				t.Fatalf("position = %v, want %v", got, tt.want)
			}
			if got := obj.SmoothPos(0.5); got != V(tt.want) {
				t.Errorf("interpolated position = %v, want %v", got, V(tt.want))
			}
			for _, s := range []*cp.Shape{shape, extra} {
				center := tt.want
				if s == extra {
					center = center.Add(offset)
				}
				if got := space.PointQueryNearest(center, 0, cp.SHAPE_FILTER_ALL).Shape; got != s {
					t.Errorf("shape at %v = %p, want %p", center, got, s)
				}
			}
			if !tt.pos.Equal(tt.want) {
				if got := space.PointQueryNearest(tt.pos, 0, cp.SHAPE_FILTER_ALL).Shape; got != nil {
					t.Errorf("shape still found at old position: %p", got)
				}
			}
		})
	}
}
