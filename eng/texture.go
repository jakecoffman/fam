package eng

import (
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"

	"github.com/go-gl/gl/v3.3-core/gl"
)

type Texture2D struct {
	ID            uint32
	Width, Height int

	InternalFormat int32
	ImageFormat    uint32

	WrapS, WrapT, FilterMin, FilterMax int32
}

func NewTexture() *Texture2D {
	var ID uint32
	gl.GenTextures(1, &ID)
	return &Texture2D{
		ID:             ID,
		InternalFormat: gl.RGBA,
		ImageFormat:    gl.RGBA,
		WrapS:          gl.REPEAT,
		WrapT:          gl.REPEAT,
		FilterMin:      gl.LINEAR,
		FilterMax:      gl.LINEAR,
	}
}

func (t *Texture2D) Generate(reader io.ReadCloser) {
	defer reader.Close()
	img, _, err := image.Decode(reader)
	if err != nil {
		log.Println("Error decoding image:", err)
		return
	}

	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Pt(0, 0), draw.Src)
	size := rgba.Rect.Size()
	t.Width = size.X
	t.Height = size.Y

	// load and create a texture
	gl.BindTexture(gl.TEXTURE_2D, t.ID) // all upcoming GL_TEXTURE_2D operations now have effect on this texture object
	// set the texture wrapping parameters
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, t.WrapS)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, t.WrapT)
	// set texture filtering parameters
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, t.FilterMin)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, t.FilterMax)
	gl.TexImage2D(gl.TEXTURE_2D, 0, t.InternalFormat, int32(size.X), int32(size.Y), 0, t.ImageFormat, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))
	// unbind
	gl.BindTexture(gl.TEXTURE_2D, 0)

	return
}

func (t *Texture2D) Bind() {
	gl.BindTexture(gl.TEXTURE_2D, t.ID)
}
