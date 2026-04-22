package sim

type RotatorTurn uint8

const (
	TurnLeft RotatorTurn = iota
	TurnRight
)

type Rotator struct {
	Turn RotatorTurn
}

func (*Rotator) Kind() ComponentKind { return KindRotator }

func (r *Rotator) Apply(s Subject) Subject {
	switch r.Turn {
	case TurnLeft:
		s.Direction = s.Direction.Left()
	case TurnRight:
		s.Direction = s.Direction.Right()
	}
	return s
}
