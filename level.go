package fam

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
)

type Level struct {
	Bricks       []*Object
	block, solid *Texture2D
}

func NewLevel(block, solid *Texture2D) *Level {
	return &Level{
		Bricks: []*Object{},
		block:  block,
		solid:  solid,
	}
}

func (l *Level) Load(file string, lvlWidth, lvlHeight int) {
	l.Bricks = l.Bricks[:0]

	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var tileData [][]int
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), " ")
		var row []int
		for _, part := range parts {
			i, err := strconv.Atoi(part)
			if err != nil {
				panic(fmt.Errorf("failed to parse level: %s", err.Error()))
			}
			row = append(row, i)
		}
		tileData = append(tileData, row)
	}
	if err := scanner.Err(); err != nil {
		panic(fmt.Errorf("failed to scan file: %s", err))
	}
	if len(tileData) > 0 {
		l.init(tileData, lvlWidth, lvlHeight)
	}
	return
}

func (l *Level) Draw(renderer *SpriteRenderer) {
	for _, tile := range l.Bricks {
		if !tile.Destroyed {
			tile.Draw(renderer, nil, 0)
		}
	}
}

func (l *Level) IsCompleted() bool {
	for _, tile := range l.Bricks {
		if !tile.IsSolid && !tile.Destroyed {
			return false
		}
	}
	return true
}

func (l *Level) init(tileData [][]int, lvlWidth, lvlHeight int) {
	height := len(tileData)
	width := len(tileData[0])
	unitWidth := lvlWidth / width
	unitHeight := lvlHeight / height

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if tileData[y][x] == 1 {
				pos := Vec2(unitWidth*x, unitHeight*y)
				size := Vec2(unitWidth, unitHeight)
				obj := NewGameObject(pos, size, l.solid)
				obj.Color = mgl32.Vec3{.8, .8, .7}
				obj.IsSolid = true
				l.Bricks = append(l.Bricks, obj)
			} else if tileData[y][x] > 1 {
				color := mgl32.Vec3{1, 1, 1}
				switch tileData[y][x] {
				case 2:
					color = mgl32.Vec3{.2, .6, 1}
				case 3:
					color = mgl32.Vec3{0, .7, 0}
				case 4:
					color = mgl32.Vec3{.8, .8, .4}
				case 5:
					color = mgl32.Vec3{1, .5, 0}
				}

				pos := Vec2(unitWidth*x, unitHeight*y)
				size := Vec2(unitWidth, unitHeight)
				obj := NewGameObject(pos, size, l.block)
				obj.Color = color
				l.Bricks = append(l.Bricks, obj)
			}
		}
	}

	return
}

func Vec2(x, y int) mgl32.Vec2 {
	return mgl32.Vec2{float32(x), float32(y)}
}
