package sim

import "particleaccelerator/internal/bignum"

// EffectiveInjectionCooldownTicks returns the current manual injection cooldown
// in ticks. The base is a 5-second cooldown at the current TickRate; injector
// rate upgrades shorten it so future injection-speed upgrades have one read
// site to affect.
func (s *GameState) EffectiveInjectionCooldownTicks() int {
	tickRate := s.TickRate
	if tickRate <= 0 {
		tickRate = DefaultTickRate
	}
	base := tickRate * DefaultInjectionCooldownSeconds
	if base <= 0 {
		base = 1
	}
	rateMul := s.Modifiers.Normalized().InjectorRateMul
	if rateMul.LTE(bignum.One()) {
		return base
	}
	return max(int(bignum.FromInt(base).Div(rateMul).Float64()), 1)
}

// HasInjector reports whether any placed cell can emit a Subject from the
// manual Inject action.
func (s *GameState) HasInjector() bool {
	if s.Grid == nil {
		return false
	}
	for y := range s.Grid.Cells {
		for x := range s.Grid.Cells[y] {
			if _, ok := s.Grid.Cells[y][x].Component.(ManualSpawner); ok {
				return true
			}
		}
	}
	return false
}

// CanInject reports whether pressing the Inject button can currently emit at
// least one Subject. It intentionally does not inspect individual Injector
// directions; off-grid Subjects are allowed and then lost by normal movement.
func (s *GameState) CanInject() bool {
	return s.InjectionCooldownRemaining <= 0 && s.CurrentLoad < s.EffectiveMaxLoad() && s.HasInjector()
}

// Inject commands every placed Injector to emit once, admitting Subjects under
// EffectiveMaxLoad in grid scan order. It returns the number of Subjects
// admitted and starts the global cooldown only after a successful admission.
func (s *GameState) Inject() int {
	if s.InjectionCooldownRemaining > 0 || s.CurrentLoad >= s.EffectiveMaxLoad() || s.Grid == nil {
		return 0
	}
	base := s.baseApplyContext()
	cap := s.EffectiveMaxLoad()
	admitted := 0
	for y := range s.Grid.Cells {
		for x := range s.Grid.Cells[y] {
			sp, ok := s.Grid.Cells[y][x].Component.(ManualSpawner)
			if !ok {
				continue
			}
			pos := Position{X: x, Y: y}
			ctx := base
			ctx.Pos = pos
			sub, fired := sp.Spawn(ctx, pos)
			if !fired || s.CurrentLoad+sub.Load > cap {
				continue
			}
			s.Grid.Subjects = append(s.Grid.Subjects, sub)
			s.CurrentLoad += sub.Load
			admitted++
		}
	}
	if admitted > 0 {
		s.InjectionCooldownRemaining = s.EffectiveInjectionCooldownTicks()
	}
	return admitted
}

func (s *GameState) advanceInjectionCooldown() {
	if s.InjectionCooldownRemaining > 0 {
		s.InjectionCooldownRemaining--
	}
}
