package game

import (
	"image/color"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/sergey121/chip8-emulator/chip8"
	"github.com/sergey121/chip8-emulator/keyboard"
)

const (
	screenWidth  = 64
	screenHeight = 32
	pixelSize    = 10 // масштабируем пиксели, чтобы они были видны
)

type Game struct {
	chip8 *chip8.Chip8

	loadMenuActive bool
	selectedROM    string
}

func NewGame() *Game {
	return &Game{
		chip8:          chip8.New(),
		loadMenuActive: true,
	}
}

func (g *Game) Width() int {
	return screenWidth * pixelSize
}

func (g *Game) Height() int {
	return screenHeight * pixelSize
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)

	if g.loadMenuActive {
		ebitenutil.DebugPrint(screen, "Выберите ROM:\n\n1 - Pong\n2 - Tetris")
		return
	}

	for y := range 32 {
		for x := range 64 {
			if g.chip8.Display[y][x] {
				drawPixel(screen, float32(x)*pixelSize, float32(y)*pixelSize, pixelSize, color.White)
			}
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.Width(), g.Height()
}

const instructionsPerFrame = 10

func (g *Game) Update() error {
	if g.loadMenuActive {
		if ebiten.IsKeyPressed(ebiten.Key1) {
			g.selectedROM = "pong.ch8"
			g.LoadROM()
		} else if ebiten.IsKeyPressed(ebiten.Key2) {
			g.selectedROM = "tetris.ch8"
			g.LoadROM()
		}
		return nil
	}

	keyboard.UpdateKeyPress(g.chip8)
	for i := 0; i < instructionsPerFrame; i++ {
		g.chip8.Update()
	}

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

func (g *Game) LoadROM() {
	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	romPath := path + "/games/" + g.selectedROM

	data, err := os.ReadFile(romPath)

	if err != nil {
		panic(err)
	}

	g.chip8.LoadROM(data)
	g.loadMenuActive = false
}
