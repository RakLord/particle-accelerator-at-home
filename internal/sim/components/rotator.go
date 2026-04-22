package components

import "particleaccelerator/internal/sim"

type Rotator struct {
	Orientation sim.Direction
}

func (*Rotator) Kind() sim.ComponentKind { return sim.KindRotator }

func (r *Rotator) Apply(s sim.Subject) (sim.Subject, bool) {
	entrySide := opposite(s.InDirection)
	a, b := r.openSides()
	if entrySide != a && entrySide != b {
		return s, true
	}
	if entrySide == a {
		s.Direction = b
	} else {
		s.Direction = a
	}
	return s, false
}

func (r *Rotator) openSides() (sim.Direction, sim.Direction) {
	return r.Orientation.Left(), r.Orientation
}

func opposite(d sim.Direction) sim.Direction {
	return (d + 2) % 4
}

func init() {
	sim.RegisterComponent(sim.KindRotator, func() sim.Component { return &Rotator{} })
}
