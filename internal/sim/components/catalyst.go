package components

import (
	"math"

	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/sim"
)

// Catalyst multiplies the Subject's Mass by a research-driven factor when the
// Subject's Element has research >= the Catalyst threshold. At exactly the
// threshold the multiplier is 1.0 (soft on-ramp, no cliff); past it the
// multiplier grows as `1 + k · log10(research - threshold + 1)`. Higher tiers
// steepen the curve via a larger k. Below the threshold the component is
// inert. See docs/features/0008-component-catalyst.md.
type Catalyst struct{}

// catalystResearchThreshold is the per-Element research level at which a
// Catalyst begins to act. Fixed across tiers — tier upgrades steepen the
// curve, they do not loosen the gate.
const catalystResearchThreshold = 25

// catalystKByTier is the log-curve coefficient at each tier. Index 0 unused;
// index N is k at tier N.
var catalystKByTier = []float64{0, 0.7, 0.95, 1.25}

func (*Catalyst) Kind() sim.ComponentKind { return sim.KindCatalyst }

func (c *Catalyst) Apply(ctx sim.ApplyContext, s sim.Subject) (sim.Subject, bool) {
	if ctx.Research == nil {
		return s, false
	}
	research := ctx.Research.Level(s.Element)
	if research < catalystResearchThreshold {
		return s, false
	}
	tier := sim.ClampTier(ctx.Tiers, sim.KindCatalyst, len(catalystKByTier)-1)
	k := catalystKByTier[tier]
	mul := 1 + k*math.Log10(float64(research-catalystResearchThreshold+1))
	s.Mass = s.Mass.Mul(bignum.FromFloat64(mul))
	return s, false
}

func init() {
	sim.RegisterComponent(sim.KindCatalyst, func() sim.Component { return &Catalyst{} })
}
