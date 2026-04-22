package sim

// Tick advances the simulation by one logical step.
// Order per tick:
//  1. Injector spawns (respecting Max Load).
//  2. Each Subject moves up to Speed cells; each entered cell applies its
//     Component, and Collector cells remove the Subject and award $USD.
func (s *GameState) Tick() {
	s.injectorSpawns()
	s.advanceSubjects()
	s.Ticks++
}

func (s *GameState) injectorSpawns() {
	g := s.Grid
	for y := range GridSize {
		for x := range GridSize {
			inj, ok := g.Cells[y][x].Component.(*Injector)
			if !ok {
				continue
			}
			inj.TickCounter++
			if inj.TickCounter < inj.SpawnInterval {
				continue
			}
			inj.TickCounter = 0
			if s.CurrentLoad+1 > s.MaxLoad {
				continue
			}
			g.Subjects = append(g.Subjects, Subject{
				Element:   inj.Element,
				Mass:      1.0,
				Speed:     1,
				Direction: inj.Direction,
				Position:  Position{X: x, Y: y},
				Load:      1,
			})
			s.CurrentLoad++
		}
	}
}

func (s *GameState) advanceSubjects() {
	g := s.Grid
	alive := g.Subjects[:0]
	for _, sub := range g.Subjects {
		collected, lost := s.stepSubject(&sub)
		if lost {
			s.CurrentLoad -= sub.Load
			continue
		}
		if collected {
			s.USD += collectValue(sub)
			s.Research[sub.Element]++
			s.CurrentLoad -= sub.Load
			continue
		}
		alive = append(alive, sub)
	}
	g.Subjects = alive
}

// stepSubject moves the Subject up to Speed cells, applying Components
// on each entered cell. Returns (collected, lost). If neither, the Subject
// is left at its new Position for the next tick.
func (s *GameState) stepSubject(sub *Subject) (collected, lost bool) {
	g := s.Grid
	steps := sub.Speed
	for range steps {
		dx, dy := sub.Direction.Step()
		nx, ny := sub.Position.X+dx, sub.Position.Y+dy
		if nx < 0 || nx >= GridSize || ny < 0 || ny >= GridSize {
			return false, true
		}
		sub.Position = Position{X: nx, Y: ny}
		cell := g.Cells[ny][nx]
		if cell.Component != nil {
			*sub = cell.Component.Apply(*sub)
		}
		if cell.IsCollector {
			return true, false
		}
	}
	return false, false
}
