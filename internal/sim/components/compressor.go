package components

import (
	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/sim"
)

// Compressor multiplies a Subject's Mass by the inverse of its displayed Speed
// when Speed is below 1.00. A Subject at Speed 0.10 collects at a 10× Mass
// multiplier (T1); at Speed 1.00 or above the component is a no-op. Tier
// scales the coefficient so higher tiers amplify the same slowdown further.
// See docs/features/0019-component-compressor.md.
type Compressor struct{}

// compressorCoefByTier is the coefficient applied to the 1/Speed ratio. Index
// 0 unused. Final multiplier is coef × (SpeedScale / Subject.Speed).
var compressorCoefByTier = []bignum.Decimal{
	bignum.Zero(),
	bignum.MustParse("1"),
	bignum.MustParse("1.5"),
	bignum.MustParse("2"),
}

func (*Compressor) Kind() sim.ComponentKind { return sim.KindCompressor }

func (c *Compressor) Apply(ctx sim.ApplyContext, s sim.Subject) (sim.Subject, bool) {
	if s.Speed <= 0 || s.Speed >= sim.SpeedScale {
		return s, false
	}
	tier := sim.ClampTier(ctx.Tiers, sim.KindCompressor, len(compressorCoefByTier)-1)
	ratio := bignum.FromInt64(int64(sim.SpeedScale)).Div(bignum.FromInt64(int64(s.Speed)))
	mul := compressorCoefByTier[tier].Mul(ratio)
	s.Mass = s.Mass.Mul(mul)
	return s, false
}

func init() {
	sim.RegisterComponent(sim.KindCompressor, func() sim.Component { return &Compressor{} })
}
