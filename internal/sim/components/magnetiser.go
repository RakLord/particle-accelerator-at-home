package components

import (
	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/sim"
)

// Magnetiser adds a tier-driven flat bonus to the Subject's Magnetism when the
// Subject is inside its speed band. See docs/features/component-magnetiser.md
// and docs/features/component-tiers.md.
type Magnetiser struct{}

// magnetiserBonusByTier is the flat Magnetism added per pass at each tier.
// Index 0 unused; index N is the bonus at tier N.
var magnetiserBonusByTier = []bignum.Decimal{
	bignum.Zero(),
	bignum.One(),
	bignum.FromInt(2),
	bignum.FromInt(3),
}

const magnetiserMinSpeed = 1

func (*Magnetiser) Kind() sim.ComponentKind { return sim.KindMagnetiser }

func (m *Magnetiser) Apply(ctx sim.ApplyContext, s sim.Subject) (sim.Subject, bool) {
	if s.Speed < magnetiserMinSpeed {
		return s, false
	}
	tier := sim.ClampTier(ctx.Tiers, sim.KindMagnetiser, len(magnetiserBonusByTier)-1)
	bonus := magnetiserBonusByTier[tier].Mul(ctx.Modifiers.MagnetiserBonusMul)
	s.Magnetism = s.Magnetism.Add(bonus)
	return s, false
}

func init() {
	sim.RegisterComponent(sim.KindMagnetiser, func() sim.Component { return &Magnetiser{} })
}
