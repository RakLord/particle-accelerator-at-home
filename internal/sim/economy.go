package sim

type Element string

const (
	ElementHydrogen Element = "hydrogen"
)

var elementMultipliers = map[Element]float64{
	ElementHydrogen: 1.0,
}

// collectValue is the $USD awarded when a Subject is collected.
// Magnetism accepted by the signature but coefficient-zero until Phase 2.
// See docs/features/value-formula.md.
func collectValue(s Subject) float64 {
	return s.Mass * float64(s.Speed) * elementMultipliers[s.Element]
}
