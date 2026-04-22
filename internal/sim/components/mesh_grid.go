package components

import "particleaccelerator/internal/sim"

const KindMeshGrid sim.ComponentKind = "mesh_grid"

// MeshGrid halves Speed (integer-divided) when the Subject is inside its speed
// band. Below the band it's inert, so a Speed=1 Subject isn't floored to 0 and
// trapped. See docs/features/component-mesh-grid.md.
type MeshGrid struct{}

const meshGridMinSpeed = 2

func (*MeshGrid) Kind() sim.ComponentKind { return KindMeshGrid }

func (*MeshGrid) Apply(s sim.Subject) sim.Subject {
	if s.Speed < meshGridMinSpeed {
		return s
	}
	s.Speed /= 2
	return s
}

func init() {
	sim.RegisterComponent(KindMeshGrid, func() sim.Component { return &MeshGrid{} })
}
