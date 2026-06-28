package game

type Direction int

const (
	DirUp Direction = iota
	DirRight
	DirDown
	DirLeft
)

type Point struct {
	X int
	Y int
}

type Snake struct {
	body []Point
	dir  Direction
	next Direction
}

func NewSnake(head Point) *Snake {
	return &Snake{
		body: []Point{
			head,
			{X: head.X - 1, Y: head.Y},
			{X: head.X - 2, Y: head.Y},
		},
		dir:  DirRight,
		next: DirRight,
	}
}

func (s *Snake) Head() Point {
	return s.body[0]
}

func (s *Snake) Body() []Point {
	return s.body
}

func (s *Snake) Turn(dir Direction) {
	if dir == s.dir || opposite(dir, s.dir) {
		return
	}
	s.next = dir
}

func (s *Snake) NextHead() Point {
	head := s.Head()
	switch s.next {
	case DirUp:
		head.Y--
	case DirRight:
		head.X++
	case DirDown:
		head.Y++
	case DirLeft:
		head.X--
	}
	return head
}

func (s *Snake) Move(grow bool) {
	s.dir = s.next
	next := s.NextHead()
	s.body = append([]Point{next}, s.body...)
	if !grow {
		s.body = s.body[:len(s.body)-1]
	}
}

func (s *Snake) HitsBody(p Point) bool {
	for _, part := range s.body {
		if part == p {
			return true
		}
	}
	return false
}

func (s *Snake) HitsBodyOnMove(p Point, grow bool) bool {
	body := s.body
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

func (s *Snake) Occupies(p Point) bool {
	return s.HitsBody(p)
}

func opposite(a, b Direction) bool {
	return (a == DirUp && b == DirDown) ||
		(a == DirDown && b == DirUp) ||
		(a == DirLeft && b == DirRight) ||
		(a == DirRight && b == DirLeft)
}
