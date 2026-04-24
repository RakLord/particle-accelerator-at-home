package components

import (
	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/sim"
)

// Magnetiser adds to the Subject's Magnetism when the Subject is inside its
// speed band. See docs/features/component-magnetiser.md.
type Magnetiser struct {
	Bonus bignum.Decimal
}

const magnetiserMinSpeed = 1

func (*Magnetiser) Kind() sim.ComponentKind { return sim.KindMagnetiser }

func (m *Magnetiser) Apply(_ sim.ApplyContext, s sim.Subject) (sim.Subject, bool) {
	if s.Speed < magnetiserMinSpeed {
		return s, false
	}
	s.Magnetism = s.Magnetism.Add(m.Bonus)
	return s, false
}

func init() {
	sim.RegisterComponent(sim.KindMagnetiser, func() sim.Component { return &Magnetiser{} })
}
