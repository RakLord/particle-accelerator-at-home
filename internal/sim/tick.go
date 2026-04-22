package sim

// SpeedDivisor is the number of StepProgress units equal to one cell of
// movement. With the default SpeedDivisor=10, a Subject with base Speed=1
// moves one cell every 10 ticks; Speed=20 moves two cells per tick.
// See docs/features/smooth-motion.md.
const SpeedDivisor = 10

// Tick advances the simulation by one logical step.
// Order per tick:
//  1. Injector spawns (respecting Max Load).
//  2. Each Subject accumulates Speed into StepProgress and advances one cell
//     per SpeedDivisor of accumulated progress. Each entered cell applies its
//     Component; Collector cells remove the Subject and award $USD.
func (s *GameState) Tick() {
	s.injectorSpawns()
	s.advanceSubjects()
	s.Ticks++
}

func (s *GameState) injectorSpawns() {
	g := s.Grid
	for y := range GridSize {
		for x := range GridSize {
			sp, ok := g.Cells[y][x].Component.(Spawner)
			if !ok {
				continue
			}
			sub, fired := sp.MaybeSpawn(Position{X: x, Y: y})
			if !fired {
				continue
			}
			if s.CurrentLoad+sub.Load > s.MaxLoad {
				continue
			}
			g.Subjects = append(g.Subjects, sub)
			s.CurrentLoad += sub.Load
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
			s.USD += collectValue(sub, s.Research[sub.Element])
			s.Research[sub.Element]++
			s.CurrentLoad -= sub.Load
			continue
		}
		alive = append(alive, sub)
	}
	g.Subjects = alive
}

// stepSubject accumulates Speed into StepProgress and advances the Subject one
// cell per SpeedDivisor of accumulated progress, applying Components on each
// entered cell. Returns (collected, lost). If neither, the Subject is left at
// its new Position for the next tick, with any leftover StepProgress < SpeedDivisor
// representing in-cell progress the renderer interpolates.
func (s *GameState) stepSubject(sub *Subject) (collected, lost bool) {
	g := s.Grid

	// Snapshot tick-start state for render-side interpolation. Path always
	// includes at least the starting cell so the renderer has a stable anchor.
	sub.PrevPosition = sub.Position
	sub.PrevInDirection = sub.InDirection
	sub.PrevStepProgress = sub.StepProgress
	sub.Path = append(sub.Path[:0], sub.Position)

	sub.StepProgress += sub.Speed
	for sub.StepProgress >= SpeedDivisor {
		sub.StepProgress -= SpeedDivisor
		dx, dy := sub.Direction.Step()
		nx, ny := sub.Position.X+dx, sub.Position.Y+dy
		if nx < 0 || nx >= GridSize || ny < 0 || ny >= GridSize {
			return false, true
		}
		// Record how we arrived at the new cell BEFORE Apply, so arrival direction
		// is preserved even if the cell's Component turns us (rotator).
		arrival := sub.Direction
		sub.Position = Position{X: nx, Y: ny}
		sub.InDirection = arrival
		cell := g.Cells[ny][nx]
		if cell.Component != nil {
			// Apply takes Subject by value; the returned copy shares the Path slice
			// header. No Apply impl may overwrite Path/motion-snapshot fields.
			*sub = cell.Component.Apply(*sub)
		}
		sub.Path = append(sub.Path, sub.Position)
		if cell.IsCollector {
			return true, false
		}
	}
	return false, false
}
