package eng

import (
	"image"
	"image/color"
	"testing"

	"github.com/go-gl/mathgl/mgl32"
)

func TestAtlasPixelsAndPadding(t *testing.T) {
	first := image.NewNRGBA(image.Rect(10, 20, 14, 22))
	for y := 20; y < 22; y++ {
		for x := 10; x < 14; x++ {
			first.SetNRGBA(x, y, color.NRGBA{uint8(x * 10), uint8(y * 5), 0, 128})
		}
	}
	second := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			second.SetRGBA(x, y, color.RGBA{0, 0, 255, 255})
		}
	}
	images := []image.Image{first, second}
	for _, repeatEdges := range []bool{false, true} {
		atlas, regions, err := packAtlas(images, 8, repeatEdges)
		if err != nil {
			t.Fatal(err)
		}
		if atlas.Bounds() != image.Rect(0, 0, 6, 8) {
			t.Fatalf("atlas bounds = %v, want (0,0)-(6,8)", atlas.Bounds())
		}
		for i, region := range regions {
			img := images[i]
			if region.Size() != img.Bounds().Size() {
				t.Errorf("region %d changed image dimensions", i)
			}
			padded := region.Inset(-1)
			for y := padded.Min.Y; y < padded.Max.Y; y++ {
				for x := padded.Min.X; x < padded.Max.X; x++ {
					sx, sy := x-region.Min.X, y-region.Min.Y
					if repeatEdges {
						sx = (sx + region.Dx()) % region.Dx()
						sy = (sy + region.Dy()) % region.Dy()
					} else {
						sx = min(max(sx, 0), region.Dx()-1)
						sy = min(max(sy, 0), region.Dy()-1)
					}
					want := color.RGBAModel.Convert(img.At(sx+img.Bounds().Min.X, sy+img.Bounds().Min.Y)).(color.RGBA)
					if got := atlas.RGBAAt(x, y); got != want {
						t.Fatalf("repeat=%v region %d pixel (%d,%d) = %v, want %v", repeatEdges, i, x, y, got, want)
					}
				}
			}
		}
	}
}

func TestAtlasRejectsImagesThatDoNotFit(t *testing.T) {
	for _, test := range []struct {
		name   string
		images []image.Image
		limit  int
	}{
		{"empty", nil, 8},
		{"zero dimensions", []image.Image{image.NewRGBA(image.Rect(0, 0, 0, 2))}, 8},
		{"oversized image", []image.Image{image.NewRGBA(image.Rect(0, 0, 8, 2))}, 8},
		{"oversized atlas", []image.Image{image.NewRGBA(image.Rect(0, 0, 4, 4)), image.NewRGBA(image.Rect(0, 0, 4, 4))}, 8},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, _, err := packAtlas(test.images, test.limit, false); err == nil {
				t.Error("invalid atlas accepted")
			}
		})
	}
}

func TestAtlasUV(t *testing.T) {
	if got, want := atlasUV(image.Rect(1, 2, 3, 4), 8, 8), (mgl32.Vec4{0.125, 0.25, 0.375, 0.5}); got != want {
		t.Errorf("UV region = %v, want %v", got, want)
	}
}
