package fam

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
)

func Vec2(x, y int) mgl32.Vec2 {
	return mgl32.Vec2{float32(x), float32(y)}
}

func V(vector cp.Vector) mgl32.Vec2 {
	return mgl32.Vec2{
		float32(vector.X),
		float32(vector.Y),
	}
}
