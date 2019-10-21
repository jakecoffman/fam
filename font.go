package fam

import (
	"fmt"
	"image"
	"image/draw"
	"io/ioutil"
	"os"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type TextRenderer struct {
	shader   *Shader
	vao, vbo uint32
	fontChar []character
	texture  uint32 // Holds the glyph texture id.
}

type character struct {
	textureID uint32 // ID handle of the glyph texture
	width     int    //glyph width
	height    int    //glyph height
	advance   int    //glyph advance
	bearingH  int    //glyph bearing horizontal
	bearingV  int    //glyph bearing vertical
}

func NewTextRenderer(shader *Shader, width, height float32, font string, scale uint32) *TextRenderer {
	shader.Use().SetMat4("projection", mgl32.Ortho2D(0, width, height, 0)).SetInt("text", 0)
	var VAO, VBO uint32
	gl.GenVertexArrays(1, &VAO)
	gl.GenBuffers(1, &VBO)
	gl.BindVertexArray(VAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, 6*4*4, nil, gl.DYNAMIC_DRAW)

	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 4, gl.FLOAT, false, 4*4, gl.PtrOffset(0))
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
	r := &TextRenderer{
		vao:    VAO,
		vbo:    VBO,
		shader: shader,
	}
	if err := r.Load(font, scale); err != nil {
		panic(err)
	}
	return r
}

func (t *TextRenderer) Load(fontPath string, scale uint32) error {
	fd, err := os.Open(fontPath)
	if err != nil {
		return err
	}
	defer func() { _ = fd.Close() }()

	low := rune(32)
	high := rune(127)

	data, err := ioutil.ReadAll(fd)
	if err != nil {
		return err
	}

	ttf, err := truetype.Parse(data)
	if err != nil {
		return err
	}

	t.fontChar = make([]character, 0, high-low+1)
	t.SetColor(1.0, 1.0, 1.0, 1.0)

	for ch := low; ch <= high; ch++ {
		var char character

		ttfFace := truetype.NewFace(ttf, &truetype.Options{
			Size:    float64(scale),
			DPI:     72,
			Hinting: font.HintingFull,
		})

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

		fg, bg := image.White, image.Black
		rect := image.Rect(0, 0, int(gw), int(gh))
		rgba := image.NewRGBA(rect)
		draw.Draw(rgba, rgba.Bounds(), bg, image.Point{}, draw.Src)

		c := freetype.NewContext()
		c.SetDPI(72)
		c.SetFont(ttf)
		c.SetFontSize(float64(scale))
		c.SetClip(rgba.Bounds())
		c.SetDst(rgba)
		c.SetSrc(fg)
		c.SetHinting(font.HintingFull)

		px := 0 - (int(gBnd.Min.X) >> 6)
		py := gAscent
		pt := freetype.Pt(px, py)

		_, err = c.DrawString(string(ch), pt)
		if err != nil {
			return err
		}

		var texture uint32
		gl.GenTextures(1, &texture)
		gl.BindTexture(gl.TEXTURE_2D, texture)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(rgba.Rect.Dx()), int32(rgba.Rect.Dy()), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)

		char.textureID = texture

		t.fontChar = append(t.fontChar, char)
	}
	gl.BindTexture(gl.TEXTURE_2D, 0)
	return nil
}

//SetColor allows you to set the text color to be used when you draw the text
func (t *TextRenderer) SetColor(red float32, green float32, blue float32, alpha float32) {
	t.shader.Use().SetVec4f("textColor", mgl32.Vec4{red, green, blue, alpha})
}

//Printf draws a string to the screen, takes a list of arguments like printf
func (t *TextRenderer) Print(text string, x64, y64 float64, scale float32) {
	x, y := float32(x64), float32(y64)
	indices := []rune(text)
	if len(indices) == 0 {
		return
	}
	t.shader.Use()

	lowChar := rune(32)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindVertexArray(t.vao)

	for i := range indices {
		runeIndex := indices[i]

		if int(runeIndex)-int(lowChar) > len(t.fontChar) || runeIndex < lowChar {
			continue
		}

		ch := t.fontChar[runeIndex-lowChar]

		xpos := x + float32(ch.bearingH)*scale
		ypos := y - float32(ch.height-ch.bearingV)*scale
		w := float32(ch.width) * scale
		h := float32(ch.height) * scale

		var vertices = []float32{
			xpos, ypos + h, 0.0, 1.0,
			xpos + w, ypos, 1.0, 0.0,
			xpos, ypos, 0.0, 0.0,
			xpos, ypos + h, 0.0, 1.0,
			xpos + w, ypos + h, 1.0, 1.0,
			xpos + w, ypos, 1.0, 0.0,
		}

		gl.BindTexture(gl.TEXTURE_2D, ch.textureID)
		gl.BindBuffer(gl.ARRAY_BUFFER, t.vbo)
		gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(vertices)*4, gl.Ptr(vertices))

		gl.BindBuffer(gl.ARRAY_BUFFER, 0)
		gl.DrawArrays(gl.TRIANGLES, 0, 6)

		x += float32(ch.advance>>6) * scale
	}
	gl.BindVertexArray(0)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.UseProgram(0)
	return
}
