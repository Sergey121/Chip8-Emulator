package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/sergey121/chip8-emulator/chip8"
)

const (
	screenWidth  = 64
	screenHeight = 32
	pixelSize    = 10 // масштабируем пиксели, чтобы они были видны
)

type Game struct {
	chip8 *chip8.Chip8
}

func NewGame() *Game {
	return &Game{
		chip8: chip8.New(),
	}
}

func (g *Game) Width() int {
	return screenWidth * pixelSize
}

func (g *Game) Height() int {
	return screenHeight * pixelSize
}

func (g *Game) Draw(screen *ebiten.Image) {
	for y := 0; y < 32; y++ {
		for x := 0; x < 64; x++ {
			if g.chip8.Display[y][x] {
				drawPixel(screen, float32(x)*pixelSize, float32(y)*pixelSize, pixelSize, color.White)
			}
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.Width(), g.Height()
}

func (g *Game) Update() error {
	return nil
}

func drawPixel(screen *ebiten.Image, x, y, size float32, clr color.Color) {
	var path vector.Path
	path.MoveTo(x, y)
	path.LineTo(x+size, y)
	path.LineTo(x+size, y+size)
	path.LineTo(x, y+size)
	path.Close()

	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)

	img := ebiten.NewImage(1, 1)
	img.Fill(clr)

	op := &ebiten.DrawTrianglesOptions{}
	op.FillRule = ebiten.FillRuleEvenOdd

	screen.DrawTriangles(vs, is, img, op)
}
