package fam

import (
	"github.com/go-gl/mathgl/mgl32"
	"log"
)

var (
	White = mgl32.Vec3{1, 1, 1}
	Black = mgl32.Vec3{0, 0, 0}
	Grey = mgl32.Vec3{128/256., 128/256., 128/256.}

	Magenta = mgl32.Vec3{240/256., 50/256., 230/256.}

	Purple = mgl32.Vec3{145/256., 30/256., 180/256.}
	Lavender = mgl32.Vec3{230/256., 190/256., 1}

	Navy = mgl32.Vec3{0, 0, .5}
	Blue = mgl32.Vec3{0, 130/256., 200/256.}

	Teal = mgl32.Vec3{0, .5, .5}
	Cyan = mgl32.Vec3{70/256., 240/256., 240/256.}

	Green = mgl32.Vec3{60/256., 180/256., 75/256.}
	Mint = mgl32.Vec3{170/256., 1, 195/256.}

	Lime = mgl32.Vec3{210/256., 245/256., 60/256.}

	Olive = mgl32.Vec3{.5, .5, 0}
	Yellow = mgl32.Vec3{1, 1, 25/256.}
	Beige = mgl32.Vec3{1, 250/256., 200/256.}

	Brown = mgl32.Vec3{170/256., 110/256., 40/256.}
	Orange = mgl32.Vec3{245/256., 130/256., 48/256.}
	Apricot = mgl32.Vec3{1, 215/256., 180/256.}

	Maroon = mgl32.Vec3{.5, 0, 0}
	Red = mgl32.Vec3{230/256., 25/256., 75/256.}
	Pink = mgl32.Vec3{250/256., 190/256., 190/256.}
)

func init() {
	log.Println(Magenta)
}

var Colors = []mgl32.Vec3{
	White,
	Black,
	Grey,
	Magenta,
	Purple,
	Lavender,
	Navy,
	Blue,
	Teal,
	Cyan,
	Green,
	Mint,
	Lime,
	Olive,
	Yellow,
	Beige,
	Brown,
	Orange,
	Apricot,
	Maroon,
	Red,
	Pink,
}

var colorCursor = 7

func getNextColor() mgl32.Vec3 {
	color := Colors[colorCursor]
	colorCursor++
	if colorCursor >= len(Colors) {
		colorCursor = 0
	}
	return color
}