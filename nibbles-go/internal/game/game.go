package game

import (
	"image/color"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	GridWidth      = 32
	GridHeight     = 24
	CellSize       = 24
	ApplesPerLevel = 15
	BaseTPS        = 7.0
	SpeedPerLevel  = 1.15
	MaxTPS         = 18.0
	MaxLives       = 1

	screenWidth  = GridWidth * CellSize
	screenHeight = GridHeight*CellSize + hudHeight
	hudHeight    = 72
)

type Scene int

const (
	SceneMenu Scene = iota
	ScenePlaying
	ScenePaused
	SceneLevelComplete
	SceneGameOver
)

type Game struct {
	scene Scene

	level       int
	score       int
	highScore   int
	applesEaten int

	snake    *Snake
	apple    Apple
	levelMap Level
	audio    *Audio
	rng      *rand.Rand

	tickProgress float64
	levelMessage int
	input        Input
}

func New() (*Game, error) {
	highScore, _ := LoadHighScore()
	g := &Game{
		scene:     SceneMenu,
		level:     1,
		highScore: highScore,
		audio:     NewAudio(),
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	g.startLevel(1)
	return g, nil
}

func (g *Game) Update() error {
	g.input.Update()

	switch g.scene {
	case SceneMenu:
		if g.input.StartPressed {
			g.resetRun()
			g.scene = ScenePlaying
			g.audio.Start()
		}
	case ScenePlaying:
		g.updatePlaying()
	case ScenePaused:
		if g.input.PausePressed {
			g.scene = ScenePlaying
		}
		if g.input.RestartPressed {
			g.resetRun()
			g.scene = ScenePlaying
		}
	case SceneLevelComplete:
		g.levelMessage--
		if g.levelMessage <= 0 || g.input.StartPressed {
			g.startLevel(g.level + 1)
			g.scene = ScenePlaying
		}
	case SceneGameOver:
		if g.input.RestartPressed {
			g.resetRun()
			g.scene = ScenePlaying
		}
		if g.input.MenuPressed {
			g.scene = SceneMenu
		}
	}

	return nil
}

func (g *Game) updatePlaying() {
	if g.input.PausePressed {
		g.scene = ScenePaused
		return
	}
	if g.input.RestartPressed {
		g.resetRun()
		return
	}

	g.snake.Turn(g.input.Direction)
	g.tickProgress += g.levelMap.Speed / 60.0
	for g.tickProgress >= 1 {
		g.tickProgress--
		g.step()
		if g.scene != ScenePlaying {
			return
		}
	}
}

func (g *Game) step() {
	next := g.snake.NextHead()
	grow := next == g.apple.Pos
	if !insideGrid(next) || g.levelMap.Blocked(next) || g.snake.HitsBodyOnMove(next, grow) || g.levelMap.HazardsAt(next) {
		g.gameOver()
		return
	}

	g.snake.Move(grow)

	if grow {
		g.score += 100 + g.level*25
		g.applesEaten++
		g.audio.Apple()

		if g.applesEaten >= ApplesPerLevel {
			g.levelComplete()
			return
		}
		g.spawnApple()
	}

	if g.levelMap.DisappearingApples && g.apple.Age > g.apple.Lifespan {
		g.spawnApple()
	}
	g.apple.Age++
	g.levelMap.UpdateHazards()
	for _, hazard := range g.levelMap.Hazards {
		if g.snake.Occupies(hazard.Pos) {
			g.gameOver()
			return
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 8, G: 10, B: 15, A: 255})
	g.draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) resetRun() {
	g.level = 1
	g.score = 0
	g.startLevel(1)
}

func (g *Game) startLevel(level int) {
	g.level = level
	g.applesEaten = 0
	g.tickProgress = 0
	g.levelMap = NewLevel(level)
	g.snake = NewSnake(Point{X: GridWidth / 2, Y: GridHeight / 2})
	for g.levelMap.Blocked(g.snake.Head()) {
		g.snake = NewSnake(Point{X: GridWidth/2 + g.rng.Intn(5) - 2, Y: GridHeight/2 + g.rng.Intn(5) - 2})
	}
	g.spawnApple()
}

func (g *Game) spawnApple() {
	g.apple = NewApple(g.rng, g.snake, g.levelMap)
}

func (g *Game) levelComplete() {
	g.score += 500 + g.level*100
	g.audio.Level()
	g.levelMessage = 120
	g.scene = SceneLevelComplete
	g.saveHighScore()
}

func (g *Game) gameOver() {
	g.audio.Crash()
	g.scene = SceneGameOver
	g.saveHighScore()
}

func (g *Game) saveHighScore() {
	if g.score <= g.highScore {
		return
	}
	g.highScore = g.score
	_ = SaveHighScore(g.highScore)
}

func insideGrid(p Point) bool {
	return p.X >= 0 && p.X < GridWidth && p.Y >= 0 && p.Y < GridHeight
}
