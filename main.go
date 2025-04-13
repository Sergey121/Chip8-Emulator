package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sergey121/chip8-emulator/game"
)

func main() {
	game := game.NewGame()

	ebiten.SetWindowSize(game.Width(), game.Height())
	ebiten.SetWindowTitle("Chip8 Emulator")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
