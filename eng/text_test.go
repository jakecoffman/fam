package eng

import (
	"testing"

	"github.com/go-gl/mathgl/mgl32"
)

func testTextRenderer() *TextRenderer {
	renderer := &TextRenderer{fontChar: make([]character, 96)}
	renderer.fontChar['A'-32] = character{
		width: 10, height: 20, advance: 12 << 6, bearingH: 1, bearingV: 2,
		uvRect: mgl32.Vec4{0.1, 0.2, 0.3, 0.4},
	}
	renderer.fontChar['B'-32] = character{
		width: 8, height: 10, advance: 9 << 6, bearingH: -1, bearingV: 1,
		uvRect: mgl32.Vec4{0.4, 0.2, 0.5, 0.3},
	}
	return renderer
}

func TestTextBatchedLayout(t *testing.T) {
	renderer := testTextRenderer()
	if !renderer.layout("AB", 100, 50, 2) {
		t.Fatal("first layout did not generate geometry")
	}
	if len(renderer.vertices) != 2*6*4 {
		t.Fatalf("got %d vertex components, want 48", len(renderer.vertices))
	}
	for i, want := range []float32{102, 54, 0.1, 0.4, 122, 14, 0.3, 0.2} {
		if renderer.vertices[i] != want {
			t.Errorf("first glyph component %d = %v, want %v", i, renderer.vertices[i], want)
		}
	}
	for i, want := range []float32{122, 52, 0.4, 0.3} {
		if renderer.vertices[24+i] != want {
			t.Errorf("second glyph component %d = %v, want %v", i, renderer.vertices[24+i], want)
		}
	}
	first := &renderer.vertices[0]
	if renderer.layout("AB", 100, 50, 2) || &renderer.vertices[0] != first {
		t.Error("unchanged text rebuilt its geometry")
	}
	if allocations := testing.AllocsPerRun(100, func() { renderer.layout("AB", 100, 50, 2) }); allocations != 0 {
		t.Errorf("cached text layout allocated %v times", allocations)
	}
}

func TestTextLayoutInvalidation(t *testing.T) {
	renderer := testTextRenderer()
	for _, key := range []textLayout{
		{"AB", 0, 0, 1},
		{"A", 0, 0, 1},
		{"A", 1, 0, 1},
		{"A", 1, 1, 1},
		{"A", 1, 1, 2},
	} {
		if !renderer.layout(key.text, key.x, key.y, key.scale) {
			t.Errorf("changed layout %+v was not rebuilt", key)
		}
	}
	renderer.layoutValid = false
	if !renderer.layout("A", 1, 1, 2) {
		t.Error("font invalidation did not rebuild geometry")
	}
}

func TestTextUnsupportedCharactersAndEmptyString(t *testing.T) {
	renderer := testTextRenderer()
	renderer.layout("A\u0080\n\u4e16\u754cB", 0, 0, 1)
	if len(renderer.vertices) != 2*6*4 {
		t.Error("unsupported characters changed the supported glyph count")
	}
	renderer.layout("", 0, 0, 1)
	if len(renderer.vertices) != 0 {
		t.Error("empty text retained old vertices")
	}
}
