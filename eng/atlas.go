package eng

import (
	"fmt"
	"image"
	"image/draw"

	"github.com/go-gl/mathgl/mgl32"
)

// packAtlas pads images for linear filtering, preserving repeating or clamped edges.
func packAtlas(images []image.Image, maxSize int, repeatEdges bool) (*image.RGBA, []image.Rectangle, error) {
	if len(images) == 0 {
		return nil, nil, fmt.Errorf("texture atlas requires at least one image")
	}
	regions := make([]image.Rectangle, len(images))
	x, y, rowHeight, width := 0, 0, 0, 0
	for i, img := range images {
		size := img.Bounds().Size()
		w, h := size.X+2, size.Y+2
		if size.X <= 0 || size.Y <= 0 || w > maxSize || h > maxSize {
			return nil, nil, fmt.Errorf("atlas image %d (%dx%d) does not fit texture limit %d", i, size.X, size.Y, maxSize)
		}
		if x+w > maxSize {
			x = 0
			y += rowHeight
			rowHeight = 0
		}
		if y+h > maxSize {
			return nil, nil, fmt.Errorf("texture atlas exceeds texture limit %d", maxSize)
		}
		regions[i] = image.Rect(x+1, y+1, x+1+size.X, y+1+size.Y)
		x += w
		width = max(width, x)
		rowHeight = max(rowHeight, h)
	}
	atlas := image.NewRGBA(image.Rect(0, 0, width, y+rowHeight))
	for i, img := range images {
		region := regions[i]
		draw.Draw(atlas, region, img, img.Bounds().Min, draw.Src)
		left, right := region.Min.X, region.Max.X-1
		top, bottom := region.Min.Y, region.Max.Y-1
		if repeatEdges {
			left, right = right, left
			top, bottom = bottom, top
		}
		for y := region.Min.Y - 1; y <= region.Max.Y; y++ {
			sourceY := y
			if y < region.Min.Y {
				sourceY = top
			} else if y >= region.Max.Y {
				sourceY = bottom
			}
			atlas.SetRGBA(region.Min.X-1, y, atlas.RGBAAt(left, sourceY))
			atlas.SetRGBA(region.Max.X, y, atlas.RGBAAt(right, sourceY))
		}
		for x := region.Min.X; x < region.Max.X; x++ {
			atlas.SetRGBA(x, region.Min.Y-1, atlas.RGBAAt(x, top))
			atlas.SetRGBA(x, region.Max.Y, atlas.RGBAAt(x, bottom))
		}
	}
	return atlas, regions, nil
}

func atlasUV(region image.Rectangle, width, height int) mgl32.Vec4 {
	return mgl32.Vec4{
		float32(region.Min.X) / float32(width),
		float32(region.Min.Y) / float32(height),
		float32(region.Max.X) / float32(width),
		float32(region.Max.Y) / float32(height),
	}
}
