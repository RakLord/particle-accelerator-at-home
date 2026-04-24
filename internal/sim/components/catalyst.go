package components

import (
	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/sim"
)

// Catalyst multiplies the Subject's Mass by a tier-driven factor, but only
// when the Subject's Element has research ≥ the Catalyst threshold. Below the
// threshold the component is inert. See docs/features/component-catalyst.md.
type Catalyst struct{}

// catalystResearchThreshold is the per-Element research level required for a
// Catalyst to activate. Fixed across tiers — tier upgrades strengthen the
// Mass multiplier, not loosen the gate.
const catalystResearchThreshold = 25

// catalystMassMulByTier is the Mass multiplier applied when active. Index 0
// unused.
var catalystMassMulByTier = []bignum.Decimal{
	bignum.Zero(),
	bignum.MustParse("1.5"),
	bignum.MustParse("2"),
	bignum.MustParse("3"),
}

func (*Catalyst) Kind() sim.ComponentKind { return sim.KindCatalyst }

func (c *Catalyst) Apply(ctx sim.ApplyContext, s sim.Subject) (sim.Subject, bool) {
	if ctx.Research == nil || ctx.Research.Level(s.Element) < catalystResearchThreshold {
		return s, false
	}
	tier := sim.ClampTier(ctx.Tiers, sim.KindCatalyst, len(catalystMassMulByTier)-1)
	s.Mass = s.Mass.Mul(catalystMassMulByTier[tier])
	return s, false
}

func init() {
	sim.RegisterComponent(sim.KindCatalyst, func() sim.Component { return &Catalyst{} })
}
