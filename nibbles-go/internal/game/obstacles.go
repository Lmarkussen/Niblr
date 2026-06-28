package game

type Hazard struct {
	Pos Point
	A   Point
	B   Point
	dir int
}

func NewHazard(a, b Point) Hazard {
	return Hazard{Pos: a, A: a, B: b, dir: 1}
}

func (h *Hazard) Update(level *Level) {
	next := h.Pos
	if h.A.X == h.B.X {
		next.Y += h.dir
		if next.Y < min(h.A.Y, h.B.Y) || next.Y > max(h.A.Y, h.B.Y) || level.Blocked(next) {
			h.dir *= -1
			next = h.Pos
			next.Y += h.dir
		}
	} else {
		next.X += h.dir
		if next.X < min(h.A.X, h.B.X) || next.X > max(h.A.X, h.B.X) || level.Blocked(next) {
			h.dir *= -1
			next = h.Pos
			next.X += h.dir
		}
	}
	h.Pos = next
}
