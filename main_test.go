package main

import (
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func withTempAppDir(t *testing.T) {
	t.Helper()
	old := appDirOverride
	appDirOverride = t.TempDir()
	t.Cleanup(func() {
		appDirOverride = old
	})
}

func newStartedTestGame() *Game {
	g := NewGame()
	g.audio = nil
	g.difficulty = 0
	g.muted = true
	g.scores = ScoreFile{Scores: map[string][]ScoreEntry{}}
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

	g.score = 0
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

func TestReturnToMenuClearsRunButKeepsDifficultySelection(t *testing.T) {
	g := newStartedTestGame()
	g.difficulty = 2
	g.level = 4
	g.score = 1200
	g.lives = 1
	g.levelApples = 9
	g.dirQueue = []direction{up, left}

	g.returnToMenu()

	if g.state != stateMenu {
		t.Fatalf("expected menu state, got %v", g.state)
	}
	if g.difficulty != 2 {
		t.Fatalf("expected selected difficulty to remain 2, got %d", g.difficulty)
	}
	if g.score != 0 || g.level != 1 || g.levelApples != 0 {
		t.Fatalf("expected run state reset, score=%d level=%d apples=%d", g.score, g.level, g.levelApples)
	}
	if len(g.dirQueue) != 0 {
		t.Fatalf("expected input buffer cleared, got %d entries", len(g.dirQueue))
	}
}

func TestHighScoreOrderingAndTrimming(t *testing.T) {
	file := ScoreFile{Scores: map[string][]ScoreEntry{}}
	for i := 0; i < 12; i++ {
		file = addHighScore(file, ScoreEntry{
			Name:       "P" + strconv.Itoa(i),
			Score:      i * 100,
			Level:      i,
			Difficulty: "Normal",
			When:       "2026-01-" + pad2(i+1),
		})
	}

	scores := file.Scores["Normal"]
	if len(scores) != 10 {
		t.Fatalf("expected top 10 scores, got %d", len(scores))
	}
	if scores[0].Score != 1100 {
		t.Fatalf("expected highest score first, got %d", scores[0].Score)
	}
	if scores[9].Score != 200 {
		t.Fatalf("expected lowest retained score 200, got %d", scores[9].Score)
	}
}

func TestHighScoresArePerDifficulty(t *testing.T) {
	file := ScoreFile{Scores: map[string][]ScoreEntry{}}
	file = addHighScore(file, ScoreEntry{Name: "AAA", Score: 100, Level: 1, Difficulty: "Normal", When: "2026-01-01T00:00:00Z"})
	file = addHighScore(file, ScoreEntry{Name: "BBB", Score: 900, Level: 3, Difficulty: "Hard", When: "2026-01-01T00:00:00Z"})

	if len(file.Scores["Normal"]) != 1 || file.Scores["Normal"][0].Name != "AAA" {
		t.Fatal("expected normal score table to contain only normal score")
	}
	if len(file.Scores["Hard"]) != 1 || file.Scores["Hard"][0].Name != "BBB" {
		t.Fatal("expected hard score table to contain only hard score")
	}
}

func TestSettingsSaveLoad(t *testing.T) {
	withTempAppDir(t)

	if err := SaveSettings(Settings{Muted: true, Difficulty: 2}); err != nil {
		t.Fatalf("save settings: %v", err)
	}
	settings := LoadSettings()
	if !settings.Muted || settings.Difficulty != 2 {
		t.Fatalf("unexpected settings: %+v", settings)
	}
}

func TestScoresSaveLoad(t *testing.T) {
	withTempAppDir(t)
	file := ScoreFile{Scores: map[string][]ScoreEntry{
		"Insane": {
			{Name: "ACE", Score: 1200, Level: 4, Difficulty: "Insane", When: "2026-01-01T00:00:00Z"},
		},
	}}

	if err := SaveScores(file); err != nil {
		t.Fatalf("save scores: %v", err)
	}
	loaded := LoadScores()
	if len(loaded.Scores["Insane"]) != 1 || loaded.Scores["Insane"][0].Name != "ACE" {
		t.Fatalf("unexpected scores: %+v", loaded)
	}
}

func TestMissingAndCorruptFilesReturnDefaults(t *testing.T) {
	withTempAppDir(t)

	if settings := LoadSettings(); settings.Muted || settings.Difficulty != 0 {
		t.Fatalf("expected default settings for missing file, got %+v", settings)
	}
	if scores := LoadScores(); len(scores.Scores) != 0 {
		t.Fatalf("expected empty scores for missing file, got %+v", scores)
	}

	if err := os.MkdirAll(appDirOverride, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(appDirOverride, "settings.json"), []byte("{bad"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(appDirOverride, "scores.json"), []byte("{bad"), 0o644); err != nil {
		t.Fatal(err)
	}

	if settings := LoadSettings(); settings.Muted || settings.Difficulty != 0 {
		t.Fatalf("expected default settings for corrupt file, got %+v", settings)
	}
	if scores := LoadScores(); len(scores.Scores) != 0 {
		t.Fatalf("expected empty scores for corrupt file, got %+v", scores)
	}
}

func TestHighScoreQualificationAndSaveFlow(t *testing.T) {
	withTempAppDir(t)
	g := newStartedTestGame()
	g.score = 500
	g.level = 3
	g.lives = 1

	g.loseLife()
	if g.state != stateNameEntry {
		t.Fatalf("expected high score name entry, got %v", g.state)
	}
	g.nameInput = "ACE"
	g.saveCurrentHighScore()
	loaded := LoadScores()
	if len(loaded.Scores["Normal"]) != 1 {
		t.Fatalf("expected one saved normal high score, got %+v", loaded)
	}
	if loaded.Scores["Normal"][0].Name != "ACE" || loaded.Scores["Normal"][0].Level != 3 {
		t.Fatalf("unexpected saved score: %+v", loaded.Scores["Normal"][0])
	}
}

func pad2(value int) string {
	if value < 10 {
		return "0" + strconv.Itoa(value)
	}
	return strconv.Itoa(value)
}
