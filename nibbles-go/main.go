package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"nibbles-go/internal/game"
)

func main() {
	g, err := game.New()
	if err != nil {
		log.Fatal(err)
	}

	ebiten.SetWindowTitle("Nibbles Go")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
