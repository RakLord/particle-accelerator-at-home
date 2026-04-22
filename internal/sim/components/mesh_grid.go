package components

import "particleaccelerator/internal/sim"

// MeshGrid halves Speed (integer-divided) when the Subject is inside its speed
// band. Below the band it's inert, so a Speed=1 Subject isn't floored to 0 and
// trapped. See docs/features/component-mesh-grid.md.
type MeshGrid struct{}

const meshGridMinSpeed = 2

func (*MeshGrid) Kind() sim.ComponentKind { return sim.KindMeshGrid }

func (*MeshGrid) Apply(s sim.Subject) (sim.Subject, bool) {
	if s.Speed < meshGridMinSpeed {
		return s, false
	}
	s.Speed /= 2
	return s, false
}

func init() {
	sim.RegisterComponent(sim.KindMeshGrid, func() sim.Component { return &MeshGrid{} })
}
