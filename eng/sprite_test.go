package eng

import (
	"math"
	"testing"
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"
)

func TestSpriteBatchOrder(t *testing.T) {
	renderer := &SpriteRenderer{}
	a, b := &Texture2D{ID: 1}, &Texture2D{ID: 2}
	for i, texture := range []*Texture2D{a, a, b, a, a} {
		renderer.DrawSprite(texture, mgl32.Vec2{float32(i), 20}, mgl32.Vec2{10, 30}, math.Pi/2, mgl32.Vec3{float32(i), 1, 0})
	}
	if len(renderer.instances) != 5 || len(renderer.batches) != 3 {
		t.Fatalf("got %d instances and %d batches, want 5 and 3", len(renderer.instances), len(renderer.batches))
	}
	for i, want := range []spriteBatch{{a, 0, 2}, {b, 2, 1}, {a, 3, 2}} {
		if got := renderer.batches[i]; got != want {
			t.Errorf("batch %d = %+v, want %+v", i, got, want)
		}
	}
	for i, instance := range renderer.instances {
		center := instance.model.Mul4x1(mgl32.Vec4{0.5, 0.5, 0, 1})
		if !center.ApproxEqualThreshold(mgl32.Vec4{float32(i), 20, 0, 1}, 1e-5) {
			t.Errorf("sprite %d center = %v", i, center)
		}
		if instance.color != (mgl32.Vec3{float32(i), 1, 0}) {
			t.Errorf("sprite %d lost its color", i)
		}
	}
	if got := unsafe.Sizeof(spriteInstance{}); got != spriteInstanceSize {
		t.Errorf("instance size = %d, want %d", got, spriteInstanceSize)
	}
}

func TestSpriteBatchManyPlayers(t *testing.T) {
	renderer := &SpriteRenderer{}
	texture := &Texture2D{ID: 1}
	for range 1000 {
		renderer.DrawSprite(texture, mgl32.Vec2{}, mgl32.Vec2{10, 10}, 0, White)
	}
	if len(renderer.batches) != 1 || renderer.batches[0].count != 1000 {
		t.Fatal("same-texture sprites did not combine into one instanced draw batch")
	}
}

func TestSpriteAtlasBatchPreservesRegionsAndOrder(t *testing.T) {
	renderer := &SpriteRenderer{}
	a := &Texture2D{ID: 1, uvRect: mgl32.Vec4{0, 0, 0.3, 1}}
	b := &Texture2D{ID: 1, uvRect: mgl32.Vec4{0.3, 0, 0.6, 1}}
	c := &Texture2D{ID: 1, uvRect: mgl32.Vec4{0.6, 0, 1, 1}}
	textures := []*Texture2D{a, b, c, b, a}
	for i, texture := range textures {
		renderer.DrawSprite(texture, mgl32.Vec2{float32(i), 0}, mgl32.Vec2{10, 10}, 0, White)
	}
	if len(renderer.batches) != 1 || renderer.batches[0].count != len(textures) {
		t.Fatal("atlas regions did not combine into one draw batch")
	}
	for i, texture := range textures {
		if renderer.instances[i].uvRect != texture.uvRect {
			t.Errorf("sprite %d has the wrong atlas region", i)
		}
		center := renderer.instances[i].model.Mul4x1(mgl32.Vec4{0.5, 0.5, 0, 1})
		if center.X() != float32(i) {
			t.Errorf("sprite %d changed draw order or position", i)
		}
	}
}
