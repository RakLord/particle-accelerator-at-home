package components

import "particleaccelerator/internal/sim"

const KindAccelerator sim.ComponentKind = "accelerator"

type SimpleAccelerator struct {
	SpeedBonus int
}

func (*SimpleAccelerator) Kind() sim.ComponentKind { return KindAccelerator }

func (a *SimpleAccelerator) Apply(s sim.Subject) sim.Subject {
	s.Speed += a.SpeedBonus
	return s
}

func init() {
	sim.RegisterComponent(KindAccelerator, func() sim.Component { return &SimpleAccelerator{} })
}
