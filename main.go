package main

import (
	"image/color"
	"log"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	windowWidth     = 800
	windowHeight    = 600
	cellSize        = 20
	hudHeight       = 40
	gridWidth       = windowWidth / cellSize
	gridHeight      = (windowHeight - hudHeight) / cellSize
	applesPerLevel  = 15
	startLives      = 4
	startMoveFrames = 16
	minMoveFrames   = 6
	appleSpeedStep  = 2
)

type point struct {
	x int
	y int
}

type direction point

var (
	up    = direction{x: 0, y: -1}
	right = direction{x: 1, y: 0}
	down  = direction{x: 0, y: 1}
	left  = direction{x: -1, y: 0}
)

type gameState int

const (
	statePlaying gameState = iota
	statePaused
	stateLevelComplete
	stateGameOver
)

type Game struct {
	snake       []point
	dir         direction
	nextDir     direction
	apple       point
	hasApple    bool
	obstacles   map[point]bool
	rng         *rand.Rand
	state       gameState
	score       int
	level       int
	completed   int
	lives       int
	levelApples int
	moveTimer   int
	keys        map[ebiten.Key]bool
}

func NewGame() *Game {
	g := &Game{
		rng:  rand.New(rand.NewSource(time.Now().UnixNano())),
		keys: map[ebiten.Key]bool{},
	}
	g.restart()
	return g
}

func (g *Game) restart() {
	g.score = 0
	g.level = 1
	g.completed = 0
	g.lives = startLives
	g.state = statePlaying
	g.startLevel()
}

func (g *Game) startLevel() {
	g.levelApples = 0
	g.moveTimer = 0
	g.obstacles = buildObstacles(g.level)
	g.resetSnake()
	g.spawnApple()
}

func (g *Game) resetSnake() {
	center := point{x: gridWidth / 2, y: gridHeight / 2}
	g.snake = []point{
		center,
		{x: center.x - 1, y: center.y},
		{x: center.x - 2, y: center.y},
	}
	g.dir = right
	g.nextDir = right
	g.moveTimer = 0
}

func buildObstacles(level int) map[point]bool {
	obstacles := map[point]bool{}

	add := func(p point) {
		if insideGrid(p) && !nearStart(p) {
			obstacles[p] = true
		}
	}
	hline := func(y, x1, x2 int, gaps ...int) {
		for x := x1; x <= x2; x++ {
			if contains(gaps, x) {
				continue
			}
			add(point{x: x, y: y})
		}
	}
	vline := func(x, y1, y2 int, gaps ...int) {
		for y := y1; y <= y2; y++ {
			if contains(gaps, y) {
				continue
			}
			add(point{x: x, y: y})
		}
	}
	block := func(x, y, w, h int) {
		for yy := y; yy < y+h; yy++ {
			for xx := x; xx < x+w; xx++ {
				add(point{x: xx, y: yy})
			}
		}
	}

	switch clamp(level, 1, 10) {
	case 1:
	case 2:
		block(8, 7, 3, 2)
		block(29, 18, 3, 2)
	case 3:
		block(7, 6, 3, 2)
		block(30, 6, 3, 2)
		block(7, 20, 3, 2)
		block(30, 20, 3, 2)
	case 4:
		vline(10, 4, 23, 10, 18)
		vline(29, 4, 23, 9, 17)
	case 5:
		hline(7, 4, 35, 12, 13, 27, 28)
		hline(20, 4, 35, 11, 12, 26, 27)
	case 6:
		hline(8, 5, 34, 18, 19, 20, 21)
		vline(12, 5, 22, 12, 13, 14, 15)
		vline(27, 5, 22, 12, 13, 14, 15)
	case 7:
		vline(8, 4, 23, 7, 14, 21)
		vline(16, 4, 23, 6, 13, 20)
		vline(24, 4, 23, 8, 15, 22)
		vline(32, 4, 23, 5, 12, 19)
	case 8:
		hline(5, 4, 35, 9, 10, 29, 30)
		hline(11, 4, 35, 5, 6, 20, 21, 34, 35)
		hline(17, 4, 35, 13, 14, 27, 28)
		hline(23, 4, 35, 8, 9, 22, 23)
	case 9:
		vline(6, 4, 23, 9, 18)
		vline(33, 4, 23, 8, 17)
		hline(5, 6, 33, 15, 16, 25, 26)
		hline(22, 6, 33, 13, 14, 23, 24)
		block(15, 9, 2, 2)
		block(24, 16, 2, 2)
	case 10:
		vline(5, 4, 23, 7, 14, 21)
		vline(12, 4, 23, 10, 18)
		vline(28, 4, 23, 9, 17)
		vline(35, 4, 23, 6, 13, 20)
		hline(6, 5, 35, 9, 10, 22, 23, 33, 34)
		hline(14, 5, 35, 6, 7, 18, 19, 30, 31)
		hline(22, 5, 35, 11, 12, 25, 26)
	}
	return obstacles
}

func nearStart(p point) bool {
	return math.Abs(float64(p.x-gridWidth/2)) < 5 && math.Abs(float64(p.y-gridHeight/2)) < 4
}

func (g *Game) Update() error {
	g.handleInput()
	defer g.rememberKeys()

	if g.state != statePlaying {
		return nil
	}

	g.moveTimer++
	if g.moveTimer >= g.moveFrames() {
		g.moveTimer = 0
		g.step()
	}
	return nil
}

func (g *Game) handleInput() {
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		g.setDirection(up)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		g.setDirection(right)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		g.setDirection(down)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		g.setDirection(left)
	}

	if g.justPressed(ebiten.KeyP) || g.justPressed(ebiten.KeyEscape) {
		if g.state == statePlaying {
			g.state = statePaused
		} else if g.state == statePaused {
			g.state = statePlaying
		}
	}
	if g.state == stateLevelComplete && g.justPressed(ebiten.KeySpace) {
		g.continueAfterLevelComplete()
	}
	if g.state == stateGameOver && g.justPressed(ebiten.KeyR) {
		g.restart()
	}
}

func (g *Game) setDirection(dir direction) {
	if g.state != statePlaying || opposite(g.dir, dir) {
		return
	}
	g.nextDir = dir
}

func (g *Game) justPressed(key ebiten.Key) bool {
	return ebiten.IsKeyPressed(key) && !g.keys[key]
}

func (g *Game) rememberKeys() {
	for _, key := range []ebiten.Key{ebiten.KeyP, ebiten.KeyEscape, ebiten.KeyR, ebiten.KeySpace} {
		g.keys[key] = ebiten.IsKeyPressed(key)
	}
}

func (g *Game) moveFrames() int {
	frames := startMoveFrames - (g.level-1)/2 - g.levelApples*appleSpeedStep
	if frames < minMoveFrames {
		return minMoveFrames
	}
	return frames
}

func (g *Game) speedMultiplier() int {
	return g.levelApples*appleSpeedStep + 1
}

func (g *Game) step() {
	g.dir = g.nextDir
	head := g.snake[0]
	next := point{x: head.x + g.dir.x, y: head.y + g.dir.y}
	grow := g.hasApple && next == g.apple

	if g.hitsWall(next) || g.hitsObstacle(next) || g.hitsSnakeOnMove(next, grow) {
		g.loseLife()
		return
	}

	g.snake = append([]point{next}, g.snake...)
	if !grow {
		g.snake = g.snake[:len(g.snake)-1]
		return
	}

	g.score += 100
	g.levelApples++
	if g.levelApples >= applesPerLevel {
		g.completeLevel()
		return
	}
	g.spawnApple()
}

func (g *Game) completeLevel() {
	g.completed = g.level
	g.state = stateLevelComplete
}

func (g *Game) continueAfterLevelComplete() {
	if g.state != stateLevelComplete {
		return
	}
	g.level = g.completed + 1
	g.state = statePlaying
	g.startLevel()
}

func (g *Game) loseLife() {
	g.lives--
	if g.lives <= 0 {
		g.state = stateGameOver
		return
	}

	g.resetSnake()
	if !g.validApple(g.apple) {
		g.spawnApple()
	}
}

func (g *Game) hitsWall(p point) bool {
	return !insideGrid(p)
}

func (g *Game) hitsObstacle(p point) bool {
	return g.obstacles[p]
}

func (g *Game) hitsSnakeOnMove(p point, grow bool) bool {
	body := g.snake
	if !grow && len(body) > 0 {
		body = body[:len(body)-1]
	}
	for _, part := range body {
		if part == p {
			return true
		}
	}
	return false
}

func (g *Game) spawnApple() {
	open := g.reachableAppleCells()
	if len(open) == 0 {
		g.hasApple = false
		return
	}
	g.apple = open[g.rng.Intn(len(open))]
	g.hasApple = true
}

func (g *Game) reachableAppleCells() []point {
	if len(g.snake) == 0 {
		return nil
	}

	start := g.snake[0]
	if !insideGrid(start) || g.hitsObstacle(start) {
		return nil
	}

	visited := map[point]bool{start: true}
	queue := []point{start}
	open := []point{}
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]

		if p != start && !g.occupied(p) {
			open = append(open, p)
		}

		for _, dir := range []direction{up, right, down, left} {
			next := point{x: p.x + dir.x, y: p.y + dir.y}
			if visited[next] || !insideGrid(next) || g.hitsObstacle(next) || g.occupied(next) {
				continue
			}
			visited[next] = true
			queue = append(queue, next)
		}
	}
	return open
}

func (g *Game) validApple(p point) bool {
	return g.hasApple && insideGrid(p) && !g.occupied(p) && !g.hitsObstacle(p) && reachableFrom(g.snake[0], p, g.obstacles, g.snake)
}

func (g *Game) occupied(p point) bool {
	for _, part := range g.snake {
		if part == p {
			return true
		}
	}
	return false
}

func insideGrid(p point) bool {
	return p.x >= 0 && p.x < gridWidth && p.y >= 0 && p.y < gridHeight
}

func opposite(a, b direction) bool {
	return a.x+b.x == 0 && a.y+b.y == 0
}

func reachableFrom(start, target point, obstacles map[point]bool, snake []point) bool {
	if start == target {
		return true
	}
	blocked := map[point]bool{}
	for _, p := range snake {
		if p != start {
			blocked[p] = true
		}
	}

	visited := map[point]bool{start: true}
	queue := []point{start}
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		for _, dir := range []direction{up, right, down, left} {
			next := point{x: p.x + dir.x, y: p.y + dir.y}
			if next == target {
				return true
			}
			if visited[next] || !insideGrid(next) || obstacles[next] || blocked[next] {
				continue
			}
			visited[next] = true
			queue = append(queue, next)
		}
	}
	return false
}

func contains(values []int, needle int) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func clamp(value, low, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 12, G: 14, B: 18, A: 255})
	g.drawBoard(screen)
	g.drawObstacles(screen)
	g.drawApple(screen)
	g.drawSnake(screen)
	g.drawHUD(screen)

	if g.state == statePaused {
		drawCenteredText(screen, "PAUSED", "P or Esc to resume")
	}
	if g.state == stateLevelComplete {
		drawCenteredText(screen, "Level "+strconv.Itoa(g.completed)+" Complete", "Press Space to continue")
	}
	if g.state == stateGameOver {
		drawCenteredText(screen, "GAME OVER", "Score: "+strconv.Itoa(g.score)+"  Level: "+strconv.Itoa(g.level)+"  Press R to restart")
	}
}

func (g *Game) drawBoard(screen *ebiten.Image) {
	for y := 0; y < gridHeight; y++ {
		for x := 0; x < gridWidth; x++ {
			ebitenutil.DrawRect(screen, float64(x*cellSize), float64(hudHeight+y*cellSize), cellSize-1, cellSize-1, color.RGBA{R: 24, G: 28, B: 35, A: 255})
		}
	}
}

func (g *Game) drawObstacles(screen *ebiten.Image) {
	for p := range g.obstacles {
		drawCell(screen, p, color.RGBA{R: 90, G: 105, B: 130, A: 255})
	}
}

func (g *Game) drawApple(screen *ebiten.Image) {
	if g.hasApple {
		drawCell(screen, g.apple, color.RGBA{R: 225, G: 62, B: 62, A: 255})
	}
}

func (g *Game) drawSnake(screen *ebiten.Image) {
	for i, p := range g.snake {
		c := color.RGBA{R: 55, G: 205, B: 105, A: 255}
		if i == 0 {
			c = color.RGBA{R: 170, G: 240, B: 90, A: 255}
		}
		drawCell(screen, p, c)
	}
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	ebitenutil.DrawRect(screen, 0, 0, windowWidth, hudHeight, color.RGBA{R: 8, G: 10, B: 14, A: 255})
	ebitenutil.DebugPrintAt(screen, "Niblr", 12, 10)
	ebitenutil.DebugPrintAt(screen, "Score: "+strconv.Itoa(g.score), 110, 10)
	ebitenutil.DebugPrintAt(screen, "Level: "+strconv.Itoa(g.level), 245, 10)
	ebitenutil.DebugPrintAt(screen, "Apples: "+strconv.Itoa(g.levelApples)+"/"+strconv.Itoa(applesPerLevel), 365, 10)
	ebitenutil.DebugPrintAt(screen, "Speed: "+strconv.Itoa(g.speedMultiplier())+"x", 540, 10)
	ebitenutil.DebugPrintAt(screen, "Lives: "+strconv.Itoa(g.lives), 630, 10)
	ebitenutil.DebugPrintAt(screen, "P/Esc pause", 705, 10)
}

func drawCell(screen *ebiten.Image, p point, c color.Color) {
	inset := 2.0
	x := float64(p.x*cellSize) + inset
	y := float64(hudHeight+p.y*cellSize) + inset
	ebitenutil.DrawRect(screen, x, y, cellSize-inset*2, cellSize-inset*2, c)
}

func drawCenteredText(screen *ebiten.Image, title, subtitle string) {
	ebitenutil.DrawRect(screen, 0, hudHeight, windowWidth, windowHeight-hudHeight, color.RGBA{R: 0, G: 0, B: 0, A: 190})
	ebitenutil.DebugPrintAt(screen, title, windowWidth/2-len(title)*3, windowHeight/2-20)
	ebitenutil.DebugPrintAt(screen, subtitle, windowWidth/2-len(subtitle)*3, windowHeight/2+10)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return windowWidth, windowHeight
}

func main() {
	ebiten.SetWindowSize(windowWidth, windowHeight)
	ebiten.SetWindowTitle("Niblr")

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
