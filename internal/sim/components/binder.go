package components

import "particleaccelerator/internal/sim"

// Binder is a generic prestige endpoint. It destroys any incoming Subject and
// asks GameState to add that Subject to the per-Element Binder Store.
//
// Orientation is purely cosmetic — the Binder accepts Subjects from any
// direction. It exists so the player can rotate the sprite to face the
// incoming pipe, mirroring Collector's affordance.
type Binder struct {
	Orientation sim.Direction
}

func (*Binder) Kind() sim.ComponentKind { return sim.KindBinder }

func (*Binder) Apply(_ sim.ApplyContext, s sim.Subject) (sim.Subject, bool) {
	return s, true
}

func (*Binder) ApplyBank(_ sim.ApplyContext, s sim.Subject) (sim.Subject, bool, bool, sim.Element) {
	if _, ok := sim.ElementCatalog[s.Element]; !ok {
		return s, true, false, ""
	}
	return s, true, true, s.Element
}

func init() {
	sim.RegisterComponent(sim.KindBinder, func() sim.Component { return &Binder{} })
}
