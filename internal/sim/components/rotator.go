package components

import "particleaccelerator/internal/sim"

type RotatorTurn uint8

const (
	TurnLeft RotatorTurn = iota
	TurnRight
)

type Rotator struct {
	Turn RotatorTurn
}

func (*Rotator) Kind() sim.ComponentKind { return sim.KindRotator }

func (r *Rotator) Apply(s sim.Subject) sim.Subject {
	switch r.Turn {
	case TurnLeft:
		s.Direction = s.Direction.Left()
	case TurnRight:
		s.Direction = s.Direction.Right()
	}
	return s
}

func init() {
	sim.RegisterComponent(sim.KindRotator, func() sim.Component { return &Rotator{} })
}
