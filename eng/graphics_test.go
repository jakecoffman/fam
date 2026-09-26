//go:build integration

package eng

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

var graphicsWindow *glfw.Window

func init() {
	runtime.LockOSThread()
}

func TestMain(m *testing.M) {
	os.Exit(runGraphicsTests(m))
}

func runGraphicsTests(m *testing.M) int {
	if err := glfw.Init(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer glfw.Terminate()
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.Visible, glfw.False)
	var err error
	graphicsWindow, err = glfw.CreateWindow(256, 256, "Graphics tests", nil, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer graphicsWindow.Destroy()
	graphicsWindow.MakeContextCurrent()
	if err := gl.Init(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	glfw.DetachCurrentContext()
	return m.Run()
}

func graphicsContext(t *testing.T) {
	t.Helper()
	runtime.LockOSThread()
	graphicsWindow.MakeContextCurrent()
	t.Cleanup(func() {
		glfw.DetachCurrentContext()
		runtime.UnlockOSThread()
	})
	var framebuffer, texture uint32
	gl.GenFramebuffers(1, &framebuffer)
	gl.GenTextures(1, &texture)
	gl.BindFramebuffer(gl.FRAMEBUFFER, framebuffer)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA8, 256, 256, 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, texture, 0)
	t.Cleanup(func() {
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
		gl.DeleteFramebuffers(1, &framebuffer)
		gl.DeleteTextures(1, &texture)
	})
	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		t.Fatal("incomplete framebuffer")
	}
	gl.Viewport(0, 0, 256, 256)
	gl.Disable(gl.FRAMEBUFFER_SRGB)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.ClearColor(0, 0, 0, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)
}

func imageFile(t *testing.T, img image.Image) string {
	t.Helper()
	var data bytes.Buffer
	if err := png.Encode(&data, img); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "image.png")
	if err := os.WriteFile(path, data.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func checkPixel(t *testing.T, x, y int32, want [4]byte) {
	t.Helper()
	var got [4]byte
	gl.ReadPixels(x, y, 1, 1, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(&got[0]))
	if got != want {
		t.Errorf("pixel (%d,%d) = %v, want %v", x, y, got, want)
	}
}

func TestGraphicsSpriteAtlas(t *testing.T) {
	graphicsContext(t)
	resources := NewResourceManager()
	defer resources.Clear()
	red := image.NewRGBA(image.Rect(0, 0, 1, 1))
	red.SetRGBA(0, 0, color.RGBA{R: 255, A: 255})
	green := image.NewRGBA(image.Rect(0, 0, 1, 1))
	green.SetRGBA(0, 0, color.RGBA{G: 255, A: 255})
	resources.LoadTextureAtlas(map[string]string{"red": imageFile(t, red), "green": imageFile(t, green)})
	a, b := resources.Texture("red"), resources.Texture("green")
	if a.ID != b.ID {
		t.Fatal("atlas views do not share a GPU texture")
	}
	plainImage := image.NewRGBA(image.Rect(0, 0, 2, 1))
	plainImage.SetRGBA(0, 0, color.RGBA{B: 255, A: 255})
	plainImage.SetRGBA(1, 0, color.RGBA{R: 255, G: 255, A: 255})
	plain := resources.LoadTexture(imageFile(t, plainImage), "plain")
	shader := resources.LoadShader("../assets/shaders/main.vs.glsl", "../assets/shaders/main.fs.glsl", "sprite")
	shader.Use().SetInt("image", 0).SetMat4("projection", mgl32.Ortho2D(0, 256, 0, 256))
	renderer := NewSpriteRenderer(shader)
	defer renderer.Destroy()
	renderer.DrawSprite(a, mgl32.Vec2{32, 32}, mgl32.Vec2{40, 40}, 0, White)
	renderer.DrawSprite(b, mgl32.Vec2{32, 32}, mgl32.Vec2{20, 20}, 0, White)
	renderer.DrawSprite(a, mgl32.Vec2{32, 32}, mgl32.Vec2{8, 8}, 0, White)
	if len(renderer.batches) != 1 {
		t.Fatal("mixed atlas sprites did not form one batch")
	}
	renderer.DrawSprite(plain, mgl32.Vec2{96.5, 96}, mgl32.Vec2{32, 32}, 0, White)
	renderer.DrawSprite(b, mgl32.Vec2{160, 96}, mgl32.Vec2{32, 32}, 0, White)
	renderer.Flush()
	checkPixel(t, 32, 32, [4]byte{255, 0, 0, 255})
	checkPixel(t, 39, 32, [4]byte{0, 255, 0, 255})
	checkPixel(t, 51, 32, [4]byte{255, 0, 0, 255})
	checkPixel(t, 88, 96, [4]byte{0, 0, 255, 255})
	checkPixel(t, 104, 96, [4]byte{255, 255, 0, 255})
	checkPixel(t, 160, 96, [4]byte{0, 255, 0, 255})
	resources.Clear()
	if gl.IsTexture(a.ID) || gl.IsTexture(plain.ID) {
		t.Error("resource cleanup retained GPU textures")
	}
	CheckGLErrors()
}

func TestGraphicsFruitAtlasMatchesTextures(t *testing.T) {
	graphicsContext(t)
	resources := NewResourceManager()
	defer resources.Clear()
	files := map[string]string{
		"banana":     "../assets/textures/banana.png",
		"blueberry":  "../assets/textures/blueberry.png",
		"strawberry": "../assets/textures/strawberry.png",
	}
	resources.LoadTextureAtlas(files)
	names := []string{"banana", "blueberry", "strawberry"}
	for _, name := range names {
		plain := resources.LoadTexture(files[name], name+"-plain")
		atlas := resources.Texture(name)
		if atlas.Width != plain.Width || atlas.Height != plain.Height {
			t.Fatalf("%s atlas view changed dimensions", name)
		}
	}
	shader := resources.LoadShader("../assets/shaders/main.vs.glsl", "../assets/shaders/main.fs.glsl", "sprite")
	shader.Use().SetInt("image", 0).SetMat4("projection", mgl32.Ortho2D(0, 256, 256, 0))
	renderer := NewSpriteRenderer(shader)
	defer renderer.Destroy()
	var reference []byte
	for _, suffix := range []string{"-plain", ""} {
		gl.Clear(gl.COLOR_BUFFER_BIT)
		for i, name := range names {
			renderer.DrawSprite(resources.Texture(name+suffix), mgl32.Vec2{float32(80 + i*48), 128},
				mgl32.Vec2{120, 140}, float64(i-1)*0.3, White)
		}
		if suffix == "" && len(renderer.batches) != 1 {
			t.Fatal("mixed fruit did not form one batch")
		}
		renderer.Flush()
		pixels := framebufferPixels()
		if suffix != "" {
			reference = pixels
			continue
		}
		for i, value := range pixels {
			delta := int(value) - int(reference[i])
			if delta < -2 || delta > 2 {
				t.Fatalf("atlas changed fruit appearance at byte %d: got %d, want %d", i, value, reference[i])
			}
		}
	}
	CheckGLErrors()
}

func framebufferPixels() []byte {
	pixels := make([]byte, 256*256*4)
	gl.ReadPixels(0, 0, 256, 256, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(pixels))
	return pixels
}

func TestGraphicsTextBatchAndCache(t *testing.T) {
	graphicsContext(t)
	resources := NewResourceManager()
	defer resources.Clear()
	shader := resources.LoadShader("../assets/shaders/text.vs.glsl", "../assets/shaders/text.fs.glsl", "text")
	renderer := NewTextRenderer(shader, 256, 256, "../assets/fonts/Roboto-Light.ttf", 24)
	defer renderer.Destroy()
	const text = "ABC ABC"
	renderer.Print(text, 20, 70, 1)
	batched := framebufferPixels()
	lit := 0
	for i := 0; i < len(batched); i += 4 {
		if batched[i] > 0 {
			lit++
		}
	}
	if lit == 0 {
		t.Fatal("text rendered no visible glyphs")
	}
	gl.Clear(gl.COLOR_BUFFER_BIT)
	x := 20.0
	for _, ch := range text {
		renderer.Print(string(ch), x, 70, 1)
		x += float64(renderer.fontChar[ch-32].advance >> 6)
	}
	if !bytes.Equal(batched, framebufferPixels()) {
		t.Error("batched text differs from per-character rendering")
	}

	renderer.Print(text, 20, 70, 1)
	gl.BindBuffer(gl.ARRAY_BUFFER, renderer.vbo)
	marker := float32(777)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, 4, gl.Ptr(&marker))
	renderer.Print(text, 20, 70, 1)
	var got float32
	gl.GetBufferSubData(gl.ARRAY_BUFFER, 0, 4, gl.Ptr(&got))
	if got != marker {
		t.Error("unchanged text re-uploaded its vertex buffer")
	}
	renderer.Print(text, 21, 70, 1)
	gl.BindBuffer(gl.ARRAY_BUFFER, renderer.vbo)
	gl.GetBufferSubData(gl.ARRAY_BUFFER, 0, 4, gl.Ptr(&got))
	if got == marker {
		t.Error("changed text position did not re-upload geometry")
	}
	oldTexture := renderer.texture
	if err := renderer.Load("../assets/fonts/Roboto-Light.ttf", 18); err != nil {
		t.Fatal(err)
	}
	if gl.IsTexture(oldTexture) || renderer.layoutValid {
		t.Error("font reload did not release old atlas and invalidate layout")
	}
	renderer.Print(text, 21, 70, 1)
	CheckGLErrors()
}
