package game

import (
	"bytes"
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"
)

var (
	colGrid     = color.RGBA{R: 25, G: 29, B: 38, A: 255}
	colSnake    = color.RGBA{R: 54, G: 222, B: 113, A: 255}
	colHead     = color.RGBA{R: 190, G: 255, B: 108, A: 255}
	colApple    = color.RGBA{R: 236, G: 69, B: 69, A: 255}
	colWall     = color.RGBA{R: 84, G: 108, B: 145, A: 255}
	colHazard   = color.RGBA{R: 250, G: 190, B: 72, A: 255}
	colText     = color.RGBA{R: 226, G: 236, B: 226, A: 255}
	colSubtle   = color.RGBA{R: 128, G: 150, B: 158, A: 255}
	colBackdrop = color.RGBA{R: 13, G: 17, B: 24, A: 235}
	fontSource  = mustFontSource()
)

func (g *Game) draw(screen *ebiten.Image) {
	g.drawGrid(screen)
	g.drawLevel(screen)
	g.drawApple(screen)
	g.drawSnake(screen)
	g.drawHUD(screen)

	switch g.scene {
	case SceneMenu:
		g.drawOverlay(screen, "NIBBLES GO", "Enter/Space: start    Arrows/WASD: move")
	case ScenePaused:
		g.drawOverlay(screen, "PAUSED", "P/Esc: resume    R: restart")
	case SceneLevelComplete:
		g.drawOverlay(screen, fmt.Sprintf("LEVEL %d CLEAR", g.level), "Next level incoming")
	case SceneGameOver:
		g.drawOverlay(screen, "GAME OVER", "R: restart    M: main menu")
	}
}

func (g *Game) drawGrid(screen *ebiten.Image) {
	for y := 0; y < GridHeight; y++ {
		for x := 0; x < GridWidth; x++ {
			px := float64(x * CellSize)
			py := float64(y*CellSize + hudHeight)
			ebitenutil.DrawRect(screen, px, py, CellSize-1, CellSize-1, colGrid)
		}
	}
}

func (g *Game) drawLevel(screen *ebiten.Image) {
	for p := range g.levelMap.Obstacles {
		drawCell(screen, p, colWall, 2)
	}
	for _, h := range g.levelMap.Hazards {
		drawCell(screen, h.Pos, colHazard, 4)
	}
}

func (g *Game) drawApple(screen *ebiten.Image) {
	inset := 5
	if g.levelMap.DisappearingApples && g.apple.Lifespan-g.apple.Age < 120 {
		inset = 7
	}
	drawCell(screen, g.apple.Pos, colApple, inset)
}

func (g *Game) drawSnake(screen *ebiten.Image) {
	for i, p := range g.snake.Body() {
		col := colSnake
		inset := 4
		if i == 0 {
			col = colHead
			inset = 3
		}
		drawCell(screen, p, col, inset)
	}
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	ebitenutil.DrawRect(screen, 0, 0, screenWidth, hudHeight, color.RGBA{R: 10, G: 12, B: 18, A: 255})
	drawText(screen, 18, 14, 24, fmt.Sprintf("Level %d", g.level), colText)
	drawText(screen, 150, 16, 20, fmt.Sprintf("Score %06d", g.score), colText)
	drawText(screen, 338, 16, 20, fmt.Sprintf("High %06d", g.highScore), colSubtle)
	drawText(screen, 556, 16, 20, fmt.Sprintf("Apples %02d/%02d", g.applesEaten, ApplesPerLevel), colText)
	drawText(screen, 18, 48, 14, "P pause   R restart   Esc pause", colSubtle)
}

func (g *Game) drawOverlay(screen *ebiten.Image, title, subtitle string) {
	ebitenutil.DrawRect(screen, 0, hudHeight, screenWidth, screenHeight-hudHeight, colBackdrop)
	drawTextCentered(screen, screenWidth/2, screenHeight/2-34, 42, title, colText)
	drawTextCentered(screen, screenWidth/2, screenHeight/2+18, 18, subtitle, colSubtle)
}

func drawCell(screen *ebiten.Image, p Point, col color.Color, inset int) {
	x := float64(p.X*CellSize + inset)
	y := float64(p.Y*CellSize + hudHeight + inset)
	size := float64(CellSize - inset*2)
	ebitenutil.DrawRect(screen, x, y, size, size, col)
}

func drawText(screen *ebiten.Image, x, y int, size float64, msg string, col color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(col)
	text.Draw(screen, msg, textFace(size), op)
}

func drawTextCentered(screen *ebiten.Image, x, y int, size float64, msg string, col color.Color) {
	w, h := text.Measure(msg, textFace(size), 0)
	drawText(screen, x-int(w)/2, y-int(h)/2, size, msg, col)
}

func textFace(size float64) *text.GoTextFace {
	return &text.GoTextFace{Source: fontSource, Size: size}
}

func mustFontSource() *text.GoTextFaceSource {
	source, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		panic(err)
	}
	return source
}
