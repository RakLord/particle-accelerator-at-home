package components

import "particleaccelerator/internal/sim"

type SimpleAccelerator struct {
	SpeedBonus  int
	Orientation sim.Direction
}

func (*SimpleAccelerator) Kind() sim.ComponentKind { return sim.KindAccelerator }

func (a *SimpleAccelerator) Apply(s sim.Subject) (sim.Subject, bool) {
	if isVertical(a.Orientation) != isVertical(s.InDirection) {
		return s, true
	}
	s.Speed += a.SpeedBonus
	return s, false
}

func isVertical(d sim.Direction) bool {
	return d == sim.DirNorth || d == sim.DirSouth
}

func init() {
	sim.RegisterComponent(sim.KindAccelerator, func() sim.Component { return &SimpleAccelerator{} })
}
