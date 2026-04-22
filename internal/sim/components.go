package sim

import "fmt"

type ComponentKind string

// Component is the universal `(Subject) → Subject` transform that every
// placeable cell implements. Concrete Components live in the
// `internal/sim/components` subpackage and self-register via init().
type Component interface {
	Kind() ComponentKind
	Apply(s Subject) Subject
}

// Spawner is an optional capability Components can implement to emit new
// Subjects during the spawn phase of a tick. Injectors are the canonical
// implementation; any future source-type component can opt in without
// touching sim.tick.go.
type Spawner interface {
	Component
	// MaybeSpawn advances the Component's internal clock for one tick and
	// returns (subject, true) when a spawn fires this tick.
	MaybeSpawn(pos Position) (Subject, bool)
}

// ComponentFactory produces a zero-valued instance of a Component. The
// factory shape is what the save-layer uses to reconstruct concrete types
// from a persisted kind string.
type ComponentFactory func() Component

var componentRegistry = map[ComponentKind]ComponentFactory{}

// RegisterComponent wires a ComponentKind to a factory that returns an
// empty instance. Concrete components call this from init(). Duplicate
// registrations panic — a ComponentKind string is a save-format identifier
// and must be unique.
func RegisterComponent(kind ComponentKind, f ComponentFactory) {
	if _, exists := componentRegistry[kind]; exists {
		panic(fmt.Sprintf("sim: duplicate component registration for %q", kind))
	}
	componentRegistry[kind] = f
}

// RegisteredKinds returns the set of registered ComponentKinds. Intended
// for tests and diagnostics — production code should not iterate this.
func RegisteredKinds() []ComponentKind {
	out := make([]ComponentKind, 0, len(componentRegistry))
	for k := range componentRegistry {
		out = append(out, k)
	}
	return out
}

func newComponentByKind(kind ComponentKind) (Component, error) {
	f, ok := componentRegistry[kind]
	if !ok {
		return nil, fmt.Errorf("sim: unknown component kind %q", kind)
	}
	return f(), nil
}
