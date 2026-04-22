package sim

type SimpleAccelerator struct {
	SpeedBonus int
}

func (*SimpleAccelerator) Kind() ComponentKind { return KindAccelerator }

func (a *SimpleAccelerator) Apply(s Subject) Subject {
	s.Speed += a.SpeedBonus
	return s
}
