package components

import (
	"encoding/json"

	"particleaccelerator/internal/sim"
)

// MeshGrid halves Speed (integer-divided) when the Subject is inside its speed
// band. Below the band it's inert, so a Speed=1 Subject isn't floored to 0 and
// trapped. See docs/features/component-mesh-grid.md.
type MeshGrid struct {
	Orientation sim.Direction
}

const meshGridMinSpeed = 2

func (*MeshGrid) Kind() sim.ComponentKind { return sim.KindMeshGrid }

func (m *MeshGrid) Apply(_ sim.ApplyContext, s sim.Subject) (sim.Subject, bool) {
	if isVertical(m.Orientation) != isVertical(s.InDirection) {
		return s, true
	}
	if s.Speed < meshGridMinSpeed {
		return s, false
	}
	s.Speed /= 2
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
