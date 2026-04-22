package components

import "particleaccelerator/internal/sim"

const KindMagnetiser sim.ComponentKind = "magnetiser"

// Magnetiser adds to the Subject's Magnetism when the Subject is inside its
// speed band. See docs/features/component-magnetiser.md.
type Magnetiser struct {
	Bonus float64
}

const magnetiserMinSpeed = 1

func (*Magnetiser) Kind() sim.ComponentKind { return KindMagnetiser }

func (m *Magnetiser) Apply(s sim.Subject) sim.Subject {
	if s.Speed < magnetiserMinSpeed {
		return s
	}
	s.Magnetism += m.Bonus
	return s
}

func init() {
	sim.RegisterComponent(KindMagnetiser, func() sim.Component { return &Magnetiser{} })
}
