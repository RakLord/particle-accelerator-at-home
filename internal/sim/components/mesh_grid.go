package components

import (
	"encoding/json"

	"particleaccelerator/internal/sim"
)

// MeshGrid divides Speed (integer-divided) by a tier-driven divisor when the
// Subject is inside its speed band. Below the band it's inert, so a low-Speed
// Subject isn't floored to 0 and trapped.
// See docs/features/component-mesh-grid.md and docs/features/component-tiers.md.
type MeshGrid struct {
	Orientation sim.Direction
}

// meshGridDivisorByTier is the Speed divisor per tier. Index 0 unused;
// T1 halves, T2 thirds, T3 quarters. Higher tiers also raise the min-speed
// band so the floor-to-zero trap stays out of reach — see meshGridMinSpeedByTier.
var meshGridDivisorByTier = []int{0, 2, 3, 4}
var meshGridMinSpeedByTier = []int{0, 2, 3, 4}

func (*MeshGrid) Kind() sim.ComponentKind { return sim.KindMeshGrid }

func (m *MeshGrid) Apply(ctx sim.ApplyContext, s sim.Subject) (sim.Subject, bool) {
	if isVertical(m.Orientation) != isVertical(s.InDirection) {
		return s, true
	}
	tier := sim.ClampTier(ctx.Tiers, sim.KindMeshGrid, len(meshGridDivisorByTier)-1)
	if s.Speed < meshGridMinSpeedByTier[tier] {
		return s, false
	}
	s.Speed /= meshGridDivisorByTier[tier]
	return s, false
}

func (m *MeshGrid) UnmarshalJSON(data []byte) error {
	type meshGridJSON struct {
		Orientation *sim.Direction
	}
	var in meshGridJSON
	if err := json.Unmarshal(data, &in); err != nil {
		return err
	}
	if in.Orientation == nil {
		m.Orientation = sim.DirEast
		return nil
	}
	m.Orientation = *in.Orientation
	return nil
}

func init() {
	sim.RegisterComponent(sim.KindMeshGrid, func() sim.Component { return &MeshGrid{} })
}
