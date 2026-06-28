package game

import "math/rand"

type Apple struct {
	Pos      Point
	Age      int
	Lifespan int
}

func NewApple(rng *rand.Rand, snake *Snake, level Level) Apple {
	free := make([]Point, 0, GridWidth*GridHeight)
	for y := 0; y < GridHeight; y++ {
		for x := 0; x < GridWidth; x++ {
			p := Point{X: x, Y: y}
			if snake.Occupies(p) || level.Blocked(p) || level.HazardsAt(p) {
				continue
			}
			free = append(free, p)
		}
	}
	pos := Point{X: 1, Y: 1}
	if len(free) > 0 {
		pos = free[rng.Intn(len(free))]
	}

	lifespan := 999999
	if level.DisappearingApples {
		lifespan = 60 * max(4, 9-level.Number/2)
	}
	return Apple{Pos: pos, Lifespan: lifespan}
}
