package components

import (
	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/sim"
)

// Duplicator is a T-junction splitter. A Subject entering the input side is
// consumed; two copies leave via the two perpendicular output sides. Mass is
// divided per-output by a tier-driven fraction. See
// docs/features/component-duplicator.md and
// docs/adr/0009-subject-emitter-capability.md.
//
// Orientation names the INPUT side. Outputs are Orientation.Left() and
// Orientation.Right() — i.e. perpendicular to the input. A Subject arriving
// from any side other than the input is destroyed.
type Duplicator struct {
	Orientation sim.Direction
}

// duplicatorMassFracByTier is the per-output fraction of incoming Mass. Index
// 0 unused. T1 is mass-conservative (2 outputs × 0.5 = 1.0); higher tiers
// actively create mass.
var duplicatorMassFracByTier = []bignum.Decimal{
	bignum.Zero(),
	bignum.MustParse("0.5"),
	bignum.MustParse("0.6"),
	bignum.MustParse("0.75"),
}

func (*Duplicator) Kind() sim.ComponentKind { return sim.KindDuplicator }

// Apply is required by the Component interface but is never called by the
// tick loop — a Splitter-capable component is dispatched via ApplySplit. A
// defensive fallback treats direct Apply as destruction.
func (d *Duplicator) Apply(_ sim.ApplyContext, s sim.Subject) (sim.Subject, bool) {
	return s, true
}

func (d *Duplicator) ApplySplit(ctx sim.ApplyContext, s sim.Subject) (sim.Subject, []sim.Subject, bool) {
	// Entry must be from the input side. s.InDirection is the direction the
	// Subject was moving when it entered the cell; the side it came IN from
	// is the opposite. If that opposite isn't our Orientation, reject.
	entrySide := opposite(s.InDirection)
	if entrySide != d.Orientation {
		return s, nil, true
	}

	tier := sim.ClampTier(ctx.Tiers, sim.KindDuplicator, len(duplicatorMassFracByTier)-1)
	frac := duplicatorMassFracByTier[tier]
	outputMass := s.Mass.Mul(frac)

	// The two perpendicular outputs relative to the input side.
	outLeft := d.Orientation.Left()
	outRight := d.Orientation.Right()

	// Each extra starts in the emitter cell with the standard half-cell spawn
	// offset so it renders from the centre outward (matches Injector's
	// StepProgress initialisation).
	mkExtra := func(dir sim.Direction) sim.Subject {
		return sim.Subject{
			Element:      s.Element,
			Mass:         outputMass,
			Speed:        s.Speed,
			Magnetism:    s.Magnetism,
			Direction:    dir,
			InDirection:  dir,
			Position:     ctx.Pos,
			Load:         s.Load,
			StepProgress: sim.SpeedDivisor / 2,
		}
	}

	extras := []sim.Subject{mkExtra(outLeft), mkExtra(outRight)}
	return s, extras, true // input is consumed
}

func init() {
	sim.RegisterComponent(sim.KindDuplicator, func() sim.Component { return &Duplicator{} })
}
