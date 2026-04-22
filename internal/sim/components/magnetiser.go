package components

import (
	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/sim"
)

const KindMagnetiser sim.ComponentKind = "magnetiser"

// Magnetiser adds to the Subject's Magnetism when the Subject is inside its
// speed band. See docs/features/component-magnetiser.md.
type Magnetiser struct {
	Bonus bignum.Decimal
}

const magnetiserMinSpeed = 1

func (*Magnetiser) Kind() sim.ComponentKind { return KindMagnetiser }

func (m *Magnetiser) Apply(s sim.Subject) sim.Subject {
	if s.Speed < magnetiserMinSpeed {
		return s
	}
	s.Magnetism = s.Magnetism.Add(m.Bonus)
	return s
}

func init() {
	sim.RegisterComponent(KindMagnetiser, func() sim.Component { return &Magnetiser{} })
}
