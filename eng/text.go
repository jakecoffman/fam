package eng

import (
	"fmt"
	"image"
	"image/draw"
	"log"
	"os"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type TextRenderer struct {
	*Shader
	vao, vbo    uint32
	fontChar    []character
	texture     uint32
	vertices    []float32
	lastLayout  textLayout
	layoutValid bool
}

type character struct {
	uvRect   mgl32.Vec4
	width    int //glyph width
	height   int //glyph height
	advance  int //glyph advance
	bearingH int //glyph bearing horizontal
	bearingV int //glyph bearing vertical
}

type textLayout struct {
	text        string
	x, y, scale float32
}

func NewTextRenderer(shader *Shader, width, height float32, font string, scale uint32) *TextRenderer {
	shader.Use().SetMat4("projection", mgl32.Ortho2D(0, width, height, 0)).SetInt("text", 0)
	var VAO, VBO uint32
	gl.GenVertexArrays(1, &VAO)
	gl.GenBuffers(1, &VBO)
	gl.BindVertexArray(VAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)

	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 4, gl.FLOAT, false, 4*4, gl.PtrOffset(0))
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
	r := &TextRenderer{
		vao:    VAO,
		vbo:    VBO,
		Shader: shader,
	}
	if err := r.Load(font, scale); err != nil {
		panic(err)
	}
	return r
}

func (t *TextRenderer) Load(fontPath string, scale uint32) error {
	data, err := os.ReadFile(fontPath)
	if err != nil {
		return err
	}
	low := rune(32)
	high := rune(127)

	ttf, err := truetype.Parse(data)
	if err != nil {
		return err
	}

	characters := make([]character, 0, high-low+1)
	images := make([]image.Image, 0, high-low+1)
	ttfFace := truetype.NewFace(ttf, &truetype.Options{
		Size:    float64(scale),
		DPI:     72,
		Hinting: font.HintingFull,
	})
	defer func() {
		if err := ttfFace.Close(); err != nil {
			log.Printf("closing font %q: %v", fontPath, err)
		}
	}()
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(ttf)
	c.SetFontSize(float64(scale))
	c.SetSrc(image.White)
	c.SetHinting(font.HintingFull)

	for ch := low; ch <= high; ch++ {
		var char character

		gBnd, gAdv, ok := ttfFace.GlyphBounds(ch)
		if ok != true {
			return fmt.Errorf("ttf face glyphBounds error")
		}

		gh := int32((gBnd.Max.Y - gBnd.Min.Y) >> 6)
		gw := int32((gBnd.Max.X - gBnd.Min.X) >> 6)

		if gw == 0 || gh == 0 {
			gBnd = ttf.Bounds(fixed.Int26_6(scale))
			gw = int32((gBnd.Max.X - gBnd.Min.X) >> 6)
			gh = int32((gBnd.Max.Y - gBnd.Min.Y) >> 6)

			if gw == 0 || gh == 0 {
				gw = 1
				gh = 1
			}
		}

		gAscent := int(-gBnd.Min.Y) >> 6
		gdescent := int(gBnd.Max.Y) >> 6

		char.width = int(gw)
		char.height = int(gh)
		char.advance = int(gAdv)
		char.bearingV = gdescent
		char.bearingH = int(gBnd.Min.X) >> 6

		rect := image.Rect(0, 0, int(gw), int(gh))
		rgba := image.NewRGBA(rect)
		draw.Draw(rgba, rgba.Bounds(), image.Black, image.Point{}, draw.Src)

		c.SetClip(rgba.Bounds())
		c.SetDst(rgba)

		px := 0 - (int(gBnd.Min.X) >> 6)
		py := gAscent
		pt := freetype.Pt(px, py)

		_, err = c.DrawString(string(ch), pt)
		if err != nil {
			return err
		}

		characters = append(characters, char)
		images = append(images, rgba)
	}
	var maxSize int32
	gl.GetIntegerv(gl.MAX_TEXTURE_SIZE, &maxSize)
	atlas, regions, err := packAtlas(images, int(maxSize), false)
	if err != nil {
		return err
	}
	texture := NewTexture()
	texture.WrapS, texture.WrapT = gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE
	texture.generateImage(atlas)
	for i := range characters {
		characters[i].uvRect = atlasUV(regions[i], texture.Width, texture.Height)
	}
	if t.texture != 0 {
		gl.DeleteTextures(1, &t.texture)
	}
	t.texture = texture.ID
	t.fontChar = characters
	t.layoutValid = false
	t.SetColor(1, 1, 1, 1)
	return nil
}

// SetColor allows you to set the text color to be used when you draw the text
func (t *TextRenderer) SetColor(red float32, green float32, blue float32, alpha float32) {
	t.Use().SetVec4f("textColor", mgl32.Vec4{red, green, blue, alpha})
}

// Print draws one batched string, reusing its geometry when the layout is unchanged.
func (t *TextRenderer) Print(text string, x64, y64 float64, scale float32) {
	changed := t.layout(text, float32(x64), float32(y64), scale)
	if len(t.vertices) == 0 {
		return
	}
	t.Use()
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, t.texture)
	gl.BindVertexArray(t.vao)
	if changed {
		gl.BindBuffer(gl.ARRAY_BUFFER, t.vbo)
		gl.BufferData(gl.ARRAY_BUFFER, len(t.vertices)*4, gl.Ptr(t.vertices), gl.DYNAMIC_DRAW)
		gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	}
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(t.vertices)/4))
	gl.BindVertexArray(0)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.UseProgram(0)
}

func (t *TextRenderer) layout(text string, x, y, scale float32) bool {
	key := textLayout{text, x, y, scale}
	if t.layoutValid && t.lastLayout == key {
		return false
	}
	t.vertices = t.vertices[:0]
	for _, r := range text {
		index := int(r) - 32

		if index < 0 || index >= len(t.fontChar) {
			continue
		}

		ch := t.fontChar[index]

		xpos := x + float32(ch.bearingH)*scale
		ypos := y - float32(ch.height-ch.bearingV)*scale
		w := float32(ch.width) * scale
		h := float32(ch.height) * scale

		uv := ch.uvRect
		t.vertices = append(t.vertices,
			xpos, ypos+h, uv[0], uv[3],
			xpos+w, ypos, uv[2], uv[1],
			xpos, ypos, uv[0], uv[1],
			xpos, ypos+h, uv[0], uv[3],
			xpos+w, ypos+h, uv[2], uv[3],
			xpos+w, ypos, uv[2], uv[1],
		)

		x += float32(ch.advance>>6) * scale
	}
	t.lastLayout = key
	t.layoutValid = true
	return true
}

func (t *TextRenderer) Destroy() {
	gl.DeleteTextures(1, &t.texture)
	gl.DeleteVertexArrays(1, &t.vao)
	gl.DeleteBuffers(1, &t.vbo)
}
