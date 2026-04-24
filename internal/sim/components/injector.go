package components

import (
	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/sim"
)

// Injector is a source: it spawns Subjects every SpawnInterval ticks in its
// configured Direction. It implements sim.Spawner; Apply is a no-op so a
// Subject passing over an Injector is unaffected.
type Injector struct {
	Direction     sim.Direction
	SpawnInterval int
	Element       sim.Element
	TickCounter   int
}

func (*Injector) Kind() sim.ComponentKind { return sim.KindInjector }
func (*Injector) Apply(_ sim.ApplyContext, s sim.Subject) (sim.Subject, bool) {
	return s, false
}

func (inj *Injector) MaybeSpawn(_ sim.ApplyContext, pos sim.Position) (sim.Subject, bool) {
	inj.TickCounter++
	if inj.TickCounter < inj.SpawnInterval {
		return sim.Subject{}, false
	}
	inj.TickCounter = 0
	return sim.Subject{
		Element:     inj.Element,
		Mass:        bignum.One(),
		Speed:       1,
		Direction:   inj.Direction,
		InDirection: inj.Direction, // spawn cell renders as straight pass-through
		Position:    pos,
		Load:        1,
		// Start at the cell center visually (half-cell of progress already spent).
		// This costs half a SpeedDivisor off the first step but keeps spawns from
		// appearing on the cell's back edge.
		StepProgress: sim.SpeedDivisor / 2,
	}, true
}

func init() {
	sim.RegisterComponent(sim.KindInjector, func() sim.Component { return &Injector{} })
}
