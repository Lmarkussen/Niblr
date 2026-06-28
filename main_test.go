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

func TestClearingLevel30EntersVictoryStateWithoutLevel31(t *testing.T) {
	g := newStartedTestGame()
	g.level = maxDesignedLevel
	g.score = 100
	g.scores = fullHighScores("Normal", 99999)
	g.startLevel()
	g.levelApples = applesPerLevel - 1
	next := point{x: g.snake[0].x + g.dir.x, y: g.snake[0].y + g.dir.y}
	g.apple = next
	g.hasApple = true

	g.step()

	if g.state != stateVictory {
		t.Fatalf("expected victory state after level %d clear, got %v", maxDesignedLevel, g.state)
	}
	if g.level != maxDesignedLevel {
		t.Fatalf("expected level to remain %d, got %d", maxDesignedLevel, g.level)
	}

	g.continueAfterLevelComplete()
	if g.level != maxDesignedLevel {
		t.Fatalf("expected level 31 never to be created, got level %d", g.level)
	}
}

func TestVictoryStillTriggersHighScoreEntry(t *testing.T) {
	withTempAppDir(t)
	g := newStartedTestGame()
	g.level = maxDesignedLevel
	g.score = 1000
	g.startLevel()
	g.levelApples = applesPerLevel - 1
	next := point{x: g.snake[0].x + g.dir.x, y: g.snake[0].y + g.dir.y}
	g.apple = next
	g.hasApple = true

	g.step()

	if g.state != stateNameEntry {
		t.Fatalf("expected high score entry after winning with qualifying score, got %v", g.state)
	}
	g.nameInput = "WIN"
	g.saveCurrentHighScore()
	loaded := LoadScores()
	if len(loaded.Scores["Normal"]) != 1 || loaded.Scores["Normal"][0].Name != "WIN" {
		t.Fatalf("expected winning score to save, got %+v", loaded)
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
	if len(levelLayouts) != maxDesignedLevel {
		t.Fatalf("expected %d level layouts, got %d", maxDesignedLevel, len(levelLayouts))
	}

	for level := 1; level <= maxDesignedLevel; level++ {
		g := newStartedTestGame()
		g.level = level
		g.startLevel()

		cells := g.reachableAppleCells()
		if len(cells) < applesPerLevel*4 {
			t.Fatalf("level %d has too few reachable apple cells: %d", level, len(cells))
		}
		if g.obstacles[g.snake[0]] || g.obstacles[g.snake[1]] || g.obstacles[g.snake[2]] {
			t.Fatalf("level %d places obstacle on starting snake", level)
		}
		if openPercent(g.obstacles) < 55 {
			t.Fatalf("level %d is too dense: %d%% open", level, openPercent(g.obstacles))
		}
		if unreachableOpenCells(g) != 0 {
			t.Fatalf("level %d has %d unreachable open cells", level, unreachableOpenCells(g))
		}

		for i := 0; i < 100; i++ {
			g.spawnApple()
			if !g.validApple(g.apple) {
				t.Fatalf("level %d spawned invalid apple at %+v", level, g.apple)
			}
		}
	}
}

func TestLevelDifficultyBands(t *testing.T) {
	if len(buildObstacles(30)) == 0 {
		t.Fatal("expected level 30 to exist with obstacles")
	}
	if len(buildObstacles(1)) != 0 {
		t.Fatal("expected level 1 to have no obstacles")
	}
	if len(buildObstacles(2)) > 16 || len(buildObstacles(3)) > 30 {
		t.Fatalf("expected levels 2-3 to stay simple, got %d and %d obstacles", len(buildObstacles(2)), len(buildObstacles(3)))
	}
	if len(buildObstacles(4)) <= len(buildObstacles(3)) {
		t.Fatalf("expected real difficulty to start by level 4, got l3=%d l4=%d", len(buildObstacles(3)), len(buildObstacles(4)))
	}

	bands := []struct {
		from int
		to   int
	}{
		{1, 3},
		{4, 8},
		{9, 14},
		{15, 20},
		{21, 25},
		{26, 30},
	}
	last := 0.0
	for _, band := range bands {
		avg := averageObstacleCount(band.from, band.to)
		if avg <= last {
			t.Fatalf("expected obstacle average to increase by band, band %d-%d avg %.1f after %.1f", band.from, band.to, avg, last)
		}
		last = avg
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

func TestParseCLIDefaultsToMenuMode(t *testing.T) {
	options, err := parseCLI(nil)
	if err != nil {
		t.Fatalf("parse cli: %v", err)
	}
	if options.DebugStart {
		t.Fatal("expected no debug start without flags")
	}
	if options.Level != 1 || options.Lives != startLives || options.Difficulty != -1 {
		t.Fatalf("unexpected defaults: %+v", options)
	}
}

func TestParseCLIDebugOptions(t *testing.T) {
	options, err := parseCLI([]string{"--level", "12", "--difficulty", "hard", "--lives", "7", "--mute"})
	if err != nil {
		t.Fatalf("parse cli: %v", err)
	}
	if !options.DebugStart {
		t.Fatal("expected debug start")
	}
	if options.Level != 12 || options.Difficulty != 1 || options.Lives != 7 || !options.Muted {
		t.Fatalf("unexpected options: %+v", options)
	}
}

func TestParseCLIRejectsInvalidLevelBounds(t *testing.T) {
	if _, err := parseCLI([]string{"--level", "0"}); err == nil {
		t.Fatal("expected level 0 to fail")
	}
	if _, err := parseCLI([]string{"--level", strconv.Itoa(maxDesignedLevel + 1)}); err == nil {
		t.Fatal("expected level above max to fail")
	}
}

func TestParseCLIRejectsInvalidDifficulty(t *testing.T) {
	if _, err := parseCLI([]string{"--difficulty", "nightmare"}); err == nil {
		t.Fatal("expected invalid difficulty to fail")
	}
}

func TestParseCLIRejectsInvalidLives(t *testing.T) {
	if _, err := parseCLI([]string{"--lives", "0"}); err == nil {
		t.Fatal("expected zero lives to fail")
	}
}

func TestApplyCLIStartsAtRequestedLevel(t *testing.T) {
	g := NewGame()
	g.audio = nil
	g.applyCLI(cliOptions{DebugStart: true, Level: 17, Difficulty: 2, Lives: 6, Muted: true})
	if g.state != statePlaying {
		t.Fatalf("expected playing state, got %v", g.state)
	}
	if g.level != 17 || g.difficulty != 2 || g.lives != 6 || !g.muted {
		t.Fatalf("unexpected applied game state: level=%d difficulty=%d lives=%d muted=%v", g.level, g.difficulty, g.lives, g.muted)
	}
	if len(g.obstacles) == 0 || !g.hasApple {
		t.Fatal("expected requested level to be initialized")
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

func fullHighScores(difficulty string, score int) ScoreFile {
	file := ScoreFile{Scores: map[string][]ScoreEntry{}}
	for i := 0; i < 10; i++ {
		file = addHighScore(file, ScoreEntry{
			Name:       "CPU",
			Score:      score + i,
			Level:      maxDesignedLevel,
			Difficulty: difficulty,
			When:       "2026-01-" + pad2(i+1) + "T00:00:00Z",
		})
	}
	return file
}

func openPercent(obstacles map[point]bool) int {
	open := gridWidth*gridHeight - len(obstacles)
	return open * 100 / (gridWidth * gridHeight)
}

func unreachableOpenCells(g *Game) int {
	reachable := map[point]bool{}
	for _, p := range g.reachableAppleCells() {
		reachable[p] = true
	}
	if len(g.snake) > 0 {
		reachable[g.snake[0]] = true
	}

	unreachable := 0
	for y := 0; y < gridHeight; y++ {
		for x := 0; x < gridWidth; x++ {
			p := point{x: x, y: y}
			if g.obstacles[p] || g.occupied(p) {
				continue
			}
			if !reachable[p] {
				unreachable++
			}
		}
	}
	return unreachable
}

func averageObstacleCount(from, to int) float64 {
	total := 0
	for level := from; level <= to; level++ {
		total += len(buildObstacles(level))
	}
	return float64(total) / float64(to-from+1)
}

func pad2(value int) string {
	if value < 10 {
		return "0" + strconv.Itoa(value)
	}
	return strconv.Itoa(value)
}
