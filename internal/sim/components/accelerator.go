package components

import "particleaccelerator/internal/sim"

type SimpleAccelerator struct {
	SpeedBonus int
}

func (*SimpleAccelerator) Kind() sim.ComponentKind { return sim.KindAccelerator }

func (a *SimpleAccelerator) Apply(s sim.Subject) sim.Subject {
	s.Speed += a.SpeedBonus
	return s
}

func init() {
	sim.RegisterComponent(sim.KindAccelerator, func() sim.Component { return &SimpleAccelerator{} })
}
