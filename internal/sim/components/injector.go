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

func (inj *Injector) MaybeSpawn(ctx sim.ApplyContext, pos sim.Position) (sim.Subject, bool) {
	inj.TickCounter++
	effective := effectiveSpawnInterval(inj.SpawnInterval, ctx.Modifiers.InjectorRateMul)
	if inj.TickCounter < effective {
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

// effectiveSpawnInterval divides the configured SpawnInterval by the global
// InjectorRateMul, clamped to at least 1 tick. rateMul below 1 is treated as
// identity — upgrades make Injectors fire faster, never slower.
func effectiveSpawnInterval(base int, rateMul bignum.Decimal) int {
	if rateMul.LTE(bignum.One()) {
		return base
	}
	eff := int(bignum.FromInt(base).Div(rateMul).Float64())
	return max(eff, 1)
}

func init() {
	sim.RegisterComponent(sim.KindInjector, func() sim.Component { return &Injector{} })
}
