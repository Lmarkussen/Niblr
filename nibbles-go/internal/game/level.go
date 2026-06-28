package game

import "math"

type Level struct {
	Number             int
	Speed              float64
	Obstacles          map[Point]bool
	Hazards            []Hazard
	DisappearingApples bool
}

func NewLevel(number int) Level {
	speed := math.Min(MaxTPS, BaseTPS+float64(number-1)*SpeedPerLevel)
	l := Level{
		Number:    number,
		Speed:     speed,
		Obstacles: map[Point]bool{},
	}

	if number >= 2 {
		l.addBorder()
	}
	if number >= 3 {
		l.addStaticBlocks(number)
	}
	if number >= 4 {
		l.addCross()
	}
	if number >= 5 {
		l.addMaze(number)
	}
	if number >= 6 {
		l.Hazards = append(l.Hazards,
			NewHazard(Point{X: 3, Y: GridHeight / 3}, Point{X: GridWidth - 4, Y: GridHeight / 3}),
			NewHazard(Point{X: GridWidth - 4, Y: GridHeight * 2 / 3}, Point{X: 3, Y: GridHeight * 2 / 3}),
		)
	}
	if number >= 7 {
		l.DisappearingApples = true
	}

	return l
}

func (l Level) Blocked(p Point) bool {
	return l.Obstacles[p]
}

func (l Level) HazardsAt(p Point) bool {
	for _, h := range l.Hazards {
		if h.Pos == p {
			return true
		}
	}
	return false
}

func (l *Level) UpdateHazards() {
	for i := range l.Hazards {
		l.Hazards[i].Update(l)
	}
}

func (l *Level) addBorder() {
	for x := 0; x < GridWidth; x++ {
		l.Obstacles[Point{X: x, Y: 0}] = true
		l.Obstacles[Point{X: x, Y: GridHeight - 1}] = true
	}
	for y := 0; y < GridHeight; y++ {
		l.Obstacles[Point{X: 0, Y: y}] = true
		l.Obstacles[Point{X: GridWidth - 1, Y: y}] = true
	}
}

func (l *Level) addStaticBlocks(number int) {
	blocks := 4 + number
	for i := 0; i < blocks; i++ {
		x := 5 + (i*7)%22
		y := 4 + (i*5)%15
		l.addRect(x, y, 2, 2)
	}
}

func (l *Level) addCross() {
	centerX := GridWidth / 2
	centerY := GridHeight / 2
	for x := 5; x < GridWidth-5; x++ {
		if x < centerX-2 || x > centerX+2 {
			l.Obstacles[Point{X: x, Y: centerY}] = true
		}
	}
	for y := 4; y < GridHeight-4; y++ {
		if y < centerY-2 || y > centerY+2 {
			l.Obstacles[Point{X: centerX, Y: y}] = true
		}
	}
}

func (l *Level) addMaze(number int) {
	step := max(5, 9-number/2)
	for x := 4; x < GridWidth-4; x += step {
		for y := 3; y < GridHeight-3; y++ {
			if y%7 == 0 || y == GridHeight/2 {
				continue
			}
			l.Obstacles[Point{X: x, Y: y}] = true
		}
	}
	for y := 5; y < GridHeight-5; y += step {
		for x := 3; x < GridWidth-3; x++ {
			if x%8 == 0 || x == GridWidth/2 {
				continue
			}
			l.Obstacles[Point{X: x, Y: y}] = true
		}
	}
}

func (l *Level) addRect(x, y, w, h int) {
	for yy := y; yy < y+h; yy++ {
		for xx := x; xx < x+w; xx++ {
			l.Obstacles[Point{X: xx, Y: yy}] = true
		}
	}
}
