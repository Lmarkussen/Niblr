package main

import (
	"math/rand"
	"testing"
)

func TestSpawnAppleAvoidsSnakeAndObstacles(t *testing.T) {
	g := NewGame()
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
	}
}

func TestSpawnAppleHandlesFullBoard(t *testing.T) {
	g := NewGame()
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
	g := NewGame()
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
