package components

import "particleaccelerator/internal/sim"

// SimpleAccelerator adds a tier-driven flat bonus to Speed for Subjects
// entering along its axis. The speed value per tier lives in
// acceleratorSpeedByTier; index 0 is unused so ClampTier can return a tier
// index directly.
type SimpleAccelerator struct {
	Orientation sim.Direction
}

// acceleratorSpeedByTier is the per-tier Speed bonus. Index 0 unused;
// index N is the bonus at tier N. See docs/features/component-tiers.md.
var acceleratorSpeedByTier = []sim.Speed{0, sim.SpeedFromInt(1), sim.SpeedFromInt(2), sim.SpeedFromInt(3)}

func (*SimpleAccelerator) Kind() sim.ComponentKind { return sim.KindAccelerator }

func (a *SimpleAccelerator) Apply(ctx sim.ApplyContext, s sim.Subject) (sim.Subject, bool) {
	if isVertical(a.Orientation) != isVertical(s.InDirection) {
		return s, true
	}
	tier := sim.ClampTier(ctx.Tiers, sim.KindAccelerator, len(acceleratorSpeedByTier)-1)
	s.Speed += acceleratorSpeedByTier[tier] + sim.SpeedFromInt(ctx.Modifiers.AcceleratorSpeedBonus)
	return s, false
}

func isVertical(d sim.Direction) bool {
	return d == sim.DirNorth || d == sim.DirSouth
}

func init() {
	sim.RegisterComponent(sim.KindAccelerator, func() sim.Component { return &SimpleAccelerator{} })
}
