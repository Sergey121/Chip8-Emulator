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

	var aLogo = []byte{
		0x60, 0x00, // LD V0, 0
		0x61, 0x00, // LD V1, 0
		0xA2, 0x08, // LD I, 0x208
		0xD0, 0x15, // DRW V0, V1, 5 (рисуем спрайт 5 байт)
		// Спрайт буквы "A", начиная с 0x208:
		0xF0, // ****
		0x90, // *  *
		0xF0, // ****
		0x90, // *  *
		0x90, // *  *
	}

	var _ = []byte{
		0x60, 0x00, // LD V0, 0
		0x61, 0x00, // LD V1, 0
		0xA2, 0x0A, // LD I, 0x20A ← СПРАЙТ ТЕПЕРЬ ТУТ
		0xD0, 0x15, // DRW V0, V1, 5
		0x12, 0x00, // JP 0x200 (бесконечный цикл)
		// Спрайт (5 байт)
		0xF0, // ****
		0x90, // *  *
		0x90, // *  *
		0x90, // *  *
		0xF0, // ****
	}

	game.LoadROM(aLogo)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
