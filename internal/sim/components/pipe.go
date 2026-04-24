package components

import "particleaccelerator/internal/sim"

type Pipe struct {
	Orientation sim.Direction
}

func (*Pipe) Kind() sim.ComponentKind { return sim.KindPipe }

func (p *Pipe) Apply(_ sim.ApplyContext, s sim.Subject) (sim.Subject, bool) {
	entrySide := opposite(s.InDirection)
	a := p.Orientation
	b := opposite(p.Orientation)
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

func init() {
	sim.RegisterComponent(sim.KindPipe, func() sim.Component { return &Pipe{} })
}
