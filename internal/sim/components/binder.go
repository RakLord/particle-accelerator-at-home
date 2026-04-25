package components

import "particleaccelerator/internal/sim"

// Binder is a generic prestige endpoint. It destroys any incoming Subject and
// asks GameState to add that Subject to the per-Element Binder Store.
//
// Orientation marks the side of the cell where the entry pipe attaches.
// Subjects must arrive from that side (i.e. moving in the opposite direction);
// any other approach destroys the Subject without banking it.
type Binder struct {
	Orientation sim.Direction
}

func (*Binder) Kind() sim.ComponentKind { return sim.KindBinder }

func (b *Binder) Apply(_ sim.ApplyContext, s sim.Subject) (sim.Subject, bool) {
	return s, true
}

func (b *Binder) ApplyBank(_ sim.ApplyContext, s sim.Subject) (sim.Subject, bool, bool, sim.Element) {
	if opposite(s.InDirection) != b.Orientation {
		return s, true, false, ""
	}
	if _, ok := sim.ElementCatalog[s.Element]; !ok {
		return s, true, false, ""
	}
	return s, true, true, s.Element
}

func init() {
	sim.RegisterComponent(sim.KindBinder, func() sim.Component { return &Binder{} })
}
