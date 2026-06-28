package main

import (
	"math/rand"
	"testing"
)

func newStartedTestGame() *Game {
	g := NewGame()
	g.audio = nil
	g.restart()
	return g
}

func TestSpawnAppleAvoidsSnakeAndObstacles(t *testing.T) {
	g := newStartedTestGame()
	g.rng = rand.New(rand.NewSource(1))
	g.snake = []point{{x: 1, y: 1}, {x: 2, y: 1}, {x: 3, y: 1}}
	g.obstacles = map[point]bool{
		{x: 4, y: 1}: true,
		{x: 5, y: 1}: true,
	}

	for i := 0; i < 200; i++ {
		g.spawnApple()
		if !g.hasApple {
			t.Fatal("expected apple to spawn")
		}
		if !insideGrid(g.apple) {
			t.Fatalf("apple spawned outside grid: %+v", g.apple)
		}
		if g.occupied(g.apple) {
			t.Fatalf("apple spawned inside snake: %+v", g.apple)
		}
		if g.hitsObstacle(g.apple) {
			t.Fatalf("apple spawned inside obstacle: %+v", g.apple)
		}
		if !reachableFrom(g.snake[0], g.apple, g.obstacles, g.snake) {
			t.Fatalf("apple spawned in unreachable cell: %+v", g.apple)
		}
	}
}

func TestSpawnAppleHandlesFullBoard(t *testing.T) {
	g := newStartedTestGame()
	g.snake = g.snake[:0]
	g.obstacles = map[point]bool{}
	for y := 0; y < gridHeight; y++ {
		for x := 0; x < gridWidth; x++ {
			g.snake = append(g.snake, point{x: x, y: y})
		}
	}

	g.spawnApple()
	if g.hasApple {
		t.Fatal("expected no apple when board is full")
	}
}

func TestCollisionLogic(t *testing.T) {
	g := newStartedTestGame()
	g.snake = []point{{x: 5, y: 5}, {x: 4, y: 5}, {x: 3, y: 5}}
	g.obstacles = map[point]bool{{x: 7, y: 5}: true}

	if !g.hitsWall(point{x: -1, y: 5}) {
		t.Fatal("expected wall collision")
	}
	if !g.hitsObstacle(point{x: 7, y: 5}) {
		t.Fatal("expected obstacle collision")
	}
	if !g.hitsSnakeOnMove(point{x: 4, y: 5}, false) {
		t.Fatal("expected body collision")
	}
	if g.hitsSnakeOnMove(point{x: 3, y: 5}, false) {
		t.Fatal("moving into the current tail should be allowed when not growing")
	}
	if !g.hitsSnakeOnMove(point{x: 3, y: 5}, true) {
		t.Fatal("moving into the current tail should collide when growing")
	}
}

func TestLevelCompleteWaitsForSpaceBeforeAdvancing(t *testing.T) {
	g := newStartedTestGame()
	for i := 0; i < applesPerLevel; i++ {
		next := point{x: g.snake[0].x + g.dir.x, y: g.snake[0].y + g.dir.y}
		g.apple = next
		g.hasApple = true
		g.step()
	}

	if g.state != stateLevelComplete {
		t.Fatalf("expected level complete state, got %v", g.state)
	}
	if g.completed != 1 {
		t.Fatalf("expected completed level 1, got %d", g.completed)
	}
	if g.level != 1 {
		t.Fatalf("expected level not to advance before Space, got %d", g.level)
	}

	g.continueAfterLevelComplete()
	if g.state != statePlaying {
		t.Fatalf("expected playing after continue, got %v", g.state)
	}
	if g.level != 2 {
		t.Fatalf("expected level 2 after continue, got %d", g.level)
	}
	if g.levelApples != 0 {
		t.Fatalf("expected apple progress reset, got %d", g.levelApples)
	}
}

func TestObstacleCollisionLosesLife(t *testing.T) {
	g := newStartedTestGame()
	g.lives = 1
	g.snake = []point{{x: 5, y: 5}, {x: 4, y: 5}, {x: 3, y: 5}}
	g.dir = right
	g.nextDir = right
	g.obstacles = map[point]bool{{x: 6, y: 5}: true}

	g.step()
	if g.state != stateGameOver {
		t.Fatal("expected obstacle collision to end game when no lives remain")
	}
}

func TestSelfCollisionLosesLife(t *testing.T) {
	g := newStartedTestGame()
	g.lives = 1
	g.snake = []point{
		{x: 5, y: 5},
		{x: 5, y: 6},
		{x: 4, y: 6},
		{x: 4, y: 5},
		{x: 5, y: 5},
	}
	g.dir = down
	g.nextDir = down
	g.obstacles = map[point]bool{}

	g.step()
	if g.state != stateGameOver {
		t.Fatal("expected self collision to end game when no lives remain")
	}
}

func TestWallCollisionLosesLife(t *testing.T) {
	g := newStartedTestGame()
	g.lives = 1
	g.snake = []point{{x: 0, y: 5}, {x: 1, y: 5}, {x: 2, y: 5}}
	g.dir = left
	g.nextDir = left
	g.obstacles = map[point]bool{}

	g.step()
	if g.state != stateGameOver {
		t.Fatal("expected wall collision to end game when no lives remain")
	}
}

func TestSpeedIncreasesWithApplesAndResetsOnLevelClear(t *testing.T) {
	g := newStartedTestGame()

	if g.speedMultiplier() != 1 {
		t.Fatalf("expected starting speed 1x, got %dx", g.speedMultiplier())
	}

	g.levelApples = 3
	if g.speedMultiplier() != 7 {
		t.Fatalf("expected speed to increase with apples, got %dx", g.speedMultiplier())
	}
	if g.moveFrames() != startMoveFrames-3*appleSpeedStep {
		t.Fatalf("expected move frames to drop by %d per apple, got %d", appleSpeedStep, g.moveFrames())
	}

	g.level = 2
	g.startLevel()
	if g.speedMultiplier() != 1 {
		t.Fatalf("expected speed to reset after level start, got %dx", g.speedMultiplier())
	}
}

func TestLoseLifeKeepsCurrentLevelUntilLivesDepleted(t *testing.T) {
	g := newStartedTestGame()
	g.level = 2
	g.levelApples = 5
	g.score = 500
	g.lives = startLives
	g.obstacles = buildObstacles(g.level)
	g.spawnApple()

	g.loseLife()
	if g.state == stateGameOver {
		t.Fatal("did not expect game over while lives remain")
	}
	if g.lives != startLives-1 {
		t.Fatalf("expected one life lost, got %d lives", g.lives)
	}
	if g.level != 2 {
		t.Fatalf("expected to remain on level 2, got level %d", g.level)
	}
	if g.levelApples != 0 {
		t.Fatalf("expected apple progress to reset on death, got %d", g.levelApples)
	}
	if g.score != 500 {
		t.Fatalf("expected score to remain 500, got %d", g.score)
	}

	g.loseLife()
	g.loseLife()
	if g.state == stateGameOver {
		t.Fatal("did not expect game over before final life is lost")
	}
	if g.state != stateLifeLost {
		t.Fatalf("expected life lost state before final life is lost, got %v", g.state)
	}

	g.loseLife()
	if g.state != stateGameOver {
		t.Fatal("expected game over when all lives are depleted")
	}
}

func TestLevelLayoutsHaveValidAppleSpawnPositions(t *testing.T) {
	for level := 1; level <= 10; level++ {
		g := newStartedTestGame()
		g.level = level
		g.startLevel()

		cells := g.reachableAppleCells()
		if len(cells) < applesPerLevel*4 {
			t.Fatalf("level %d has too few reachable apple cells: %d", level, len(cells))
		}

		for i := 0; i < 100; i++ {
			g.spawnApple()
			if !g.validApple(g.apple) {
				t.Fatalf("level %d spawned invalid apple at %+v", level, g.apple)
			}
		}
	}
}

func TestNewGameStartsAtMenuAndDifficultyAffectsSpeed(t *testing.T) {
	g := NewGame()
	g.audio = nil
	if g.state != stateMenu {
		t.Fatalf("expected new game to start at menu, got %v", g.state)
	}

	g.difficulty = 0
	g.restart()
	normal := g.moveFrames()
	g.difficulty = 1
	hard := g.moveFrames()
	g.difficulty = 2
	insane := g.moveFrames()

	if hard >= normal {
		t.Fatalf("expected hard to be faster than normal, normal=%d hard=%d", normal, hard)
	}
	if insane >= hard {
		t.Fatalf("expected insane to be faster than hard, hard=%d insane=%d", hard, insane)
	}
}

func TestInputBufferRejectsReversalAndQueuesTurns(t *testing.T) {
	g := newStartedTestGame()

	g.setDirection(left)
	if len(g.dirQueue) != 0 {
		t.Fatal("expected immediate 180-degree reversal to be rejected")
	}

	g.setDirection(up)
	g.setDirection(left)
	if len(g.dirQueue) != 2 {
		t.Fatalf("expected two buffered turns, got %d", len(g.dirQueue))
	}

	g.step()
	if g.dir != up {
		t.Fatalf("expected first queued direction up, got %+v", g.dir)
	}
	g.step()
	if g.dir != left {
		t.Fatalf("expected second queued direction left, got %+v", g.dir)
	}
}
