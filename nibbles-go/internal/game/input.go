package game

import "github.com/hajimehoshi/ebiten/v2"

type Input struct {
	Direction      Direction
	StartPressed   bool
	PausePressed   bool
	RestartPressed bool
	MenuPressed    bool

	previous map[ebiten.Key]bool
}

func (i *Input) Update() {
	if i.previous == nil {
		i.previous = map[ebiten.Key]bool{}
		i.Direction = DirRight
	}

	i.StartPressed = i.justPressed(ebiten.KeyEnter) || i.justPressed(ebiten.KeySpace)
	i.PausePressed = i.justPressed(ebiten.KeyP) || i.justPressed(ebiten.KeyEscape)
	i.RestartPressed = i.justPressed(ebiten.KeyR)
	i.MenuPressed = i.justPressed(ebiten.KeyM)

	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		i.Direction = DirUp
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		i.Direction = DirRight
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		i.Direction = DirDown
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		i.Direction = DirLeft
	}

	for _, key := range trackedKeys {
		i.previous[key] = ebiten.IsKeyPressed(key)
	}
}

func (i *Input) justPressed(key ebiten.Key) bool {
	pressed := ebiten.IsKeyPressed(key)
	wasPressed := i.previous[key]
	return pressed && !wasPressed
}

var trackedKeys = []ebiten.Key{
	ebiten.KeyEnter,
	ebiten.KeySpace,
	ebiten.KeyP,
	ebiten.KeyEscape,
	ebiten.KeyR,
	ebiten.KeyM,
}
