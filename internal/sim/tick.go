package sim

import "particleaccelerator/internal/bignum"

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

// baseApplyContext builds the invariant portion of ApplyContext for this tick.
// Callers fill in Pos per-visit. Grid and Research views wrap live state but
// hand out copies/fresh slices so components cannot mutate through them.
// See docs/adr/0008-apply-context-and-grid-view.md.
func (s *GameState) baseApplyContext() ApplyContext {
	return ApplyContext{
		Grid:             newGridView(s.Grid),
		Tick:             s.Ticks,
		Research:         newResearchView(s.Research),
		Tiers:            newTierView(s.ComponentTiers),
		Modifiers:        s.Modifiers.Normalized(),
		Layer:            s.Layer,
		InjectionElement: s.effectiveInjectionElement(),
	}
}

func (s *GameState) injectorSpawns() {
	g := s.Grid
	base := s.baseApplyContext()
	cap := s.EffectiveMaxLoad()
	for y := range GridSize {
		for x := range GridSize {
			sp, ok := g.Cells[y][x].Component.(Spawner)
			if !ok {
				continue
			}
			pos := Position{X: x, Y: y}
			ctx := base
			ctx.Pos = pos
			sub, fired := sp.MaybeSpawn(ctx, pos)
			if !fired {
				continue
			}
			if s.CurrentLoad+sub.Load > cap {
				continue
			}
			g.Subjects = append(g.Subjects, sub)
			s.CurrentLoad += sub.Load
		}
	}
}

func (s *GameState) advanceSubjects() {
	g := s.Grid
	mods := s.Modifiers.Normalized()
	// Splitter-emitted extras are collected here and admitted after the filter
	// loop, so they don't get clobbered by g.Subjects[:0] reuse or processed
	// twice in the same tick.
	var pending []Subject
	alive := g.Subjects[:0]
	for _, sub := range g.Subjects {
		collected, lost := s.stepSubject(&sub, &pending)
		if lost {
			s.CurrentLoad -= sub.Load
			continue
		}
		if collected {
			research := s.Research[sub.Element]
			value := collectValue(sub, research, mods)
			s.USD = s.USD.Add(value)
			s.recordCollectionBestStats(sub, value)
			s.recordCollectionLog(sub, research, value)
			s.Research[sub.Element] += 1 + mods.ResearchPerCollectBonus
			s.CurrentLoad -= sub.Load
			continue
		}
		alive = append(alive, sub)
	}
	g.Subjects = alive
	// Admit pending extras under the EffectiveMaxLoad cap. Partial admission
	// matches Injector spawn semantics (ADR 0009).
	if len(pending) > 0 {
		cap := s.EffectiveMaxLoad()
		for _, e := range pending {
			if s.CurrentLoad+e.Load > cap {
				continue
			}
			g.Subjects = append(g.Subjects, e)
			s.CurrentLoad += e.Load
		}
	}
}

func (s *GameState) recordCollectionBestStats(sub Subject, value bignum.Decimal) {
	if s.BestStats == nil {
		s.BestStats = map[Element]ElementBestStats{}
	}
	stats := s.BestStats[sub.Element]
	if sub.Speed > stats.MaxSpeed {
		stats.MaxSpeed = sub.Speed
	}
	if sub.Mass.GT(stats.MaxMass) {
		stats.MaxMass = sub.Mass
	}
	if value.GT(stats.MaxCollectedValue) {
		stats.MaxCollectedValue = value
	}
	s.BestStats[sub.Element] = stats
}

func (s *GameState) recordCollectionLog(sub Subject, research int, value bignum.Decimal) {
	entry := CollectionLogEntry{
		Element:       sub.Element,
		Mass:          sub.Mass,
		Speed:         sub.Speed,
		Magnetism:     sub.Magnetism,
		ResearchLevel: research,
		Value:         value,
		Tick:          s.Ticks,
	}
	s.CollectionLog = append([]CollectionLogEntry{entry}, s.CollectionLog...)
	if len(s.CollectionLog) > MaxCollectionLogEntries {
		s.CollectionLog = s.CollectionLog[:MaxCollectionLogEntries]
	}
}

// stepSubject accumulates Speed into StepProgress and advances the Subject one
// cell per SpeedDivisor of accumulated progress, applying Components on each
// entered cell. Returns (collected, lost). If neither, the Subject is left at
// its new Position for the next tick, with any leftover StepProgress < SpeedDivisor
// representing in-cell progress the renderer interpolates.
//
// Splitter-emitted extras are appended to *pending, not to g.Subjects directly.
// The caller (advanceSubjects) admits them after its filter loop so they don't
// get clobbered by g.Subjects[:0] reuse. MaxLoad enforcement also lives with
// the admission — see ADR 0009.
func (s *GameState) stepSubject(sub *Subject, pending *[]Subject) (collected, lost bool) {
	g := s.Grid
	base := s.baseApplyContext()

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
		// is preserved even if the cell's Component turns us (elbow).
		arrival := sub.Direction
		sub.Position = Position{X: nx, Y: ny}
		sub.InDirection = arrival
		cell := g.Cells[ny][nx]
		// A Subject that enters an empty cell (no Component, not a Collector)
		// leaves the pipe network and is destroyed.
		if cell.Component == nil && !cell.IsCollector {
			return false, true
		}
		if cell.Component != nil {
			ctx := base
			ctx.Pos = sub.Position
			if sp, ok := cell.Component.(Splitter); ok {
				self, extras, destroyed := sp.ApplySplit(ctx, *sub)
				*sub = self
				*pending = append(*pending, extras...)
				if destroyed {
					return false, true
				}
			} else {
				// Apply takes Subject by value; the returned copy shares the Path slice
				// header. No Apply impl may overwrite Path/motion-snapshot fields.
				var destroyed bool
				*sub, destroyed = cell.Component.Apply(ctx, *sub)
				if destroyed {
					return false, true
				}
			}
		}
		sub.Path = append(sub.Path, sub.Position)
		if cell.IsCollector {
			return true, false
		}
	}
	return false, false
}
