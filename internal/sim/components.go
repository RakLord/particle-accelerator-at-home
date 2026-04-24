package sim

import "fmt"

type ComponentKind string

// Component is the universal cell behaviour every placeable tile implements.
// Concrete Components live in the `internal/sim/components` subpackage and
// self-register via init().
type Component interface {
	Kind() ComponentKind
	// Apply transforms the Subject after it enters the cell. When lost is true,
	// the Subject is destroyed by the Component and removed from the grid.
	//
	// ctx is a read-only view of world state. Implementations must not mutate
	// any object reachable through ctx, nor retain references across ticks.
	// See docs/adr/0008-apply-context-and-grid-view.md.
	Apply(ctx ApplyContext, s Subject) (Subject, bool)
}

// ManualSpawner is implemented by source Components that emit a Subject when
// the player presses the global Inject button. Cooldown and MaxLoad admission
// are owned by GameState.Inject, not by individual Components.
type ManualSpawner interface {
	Component
	Spawn(ctx ApplyContext, pos Position) (Subject, bool)
}

// Splitter is an optional capability for components that, on Apply, may
// produce extra Subjects in addition to transforming the incoming one.
// Components that don't emit extras implement Component, not Splitter.
//
// See docs/adr/0009-subject-emitter-capability.md.
type Splitter interface {
	Component
	// ApplySplit is the emitter-aware variant of Apply. The first return is
	// the transformed incoming Subject (same shape as Apply). Extras are
	// additional Subjects the component emits this cell visit; each is
	// appended to the grid with its Load charged against MaxLoad individually.
	// If the incoming Subject is lost, extras are still emitted — a Splitter
	// can consume the input and emit replacements.
	ApplySplit(ctx ApplyContext, s Subject) (self Subject, extras []Subject, lost bool)
}

// ApplyContext is the read-only view of world state handed to components
// during a tick. All fields are safe to read; implementations must not
// mutate any referenced object nor retain references across ticks.
//
// Modifiers is guaranteed to be Normalized when the tick loop constructs the
// context — Decimal fields can be multiplied without zero-guards. Tests that
// build ApplyContext by hand should pass Modifiers through Normalized() too,
// or use NewTestApplyContext below.
type ApplyContext struct {
	Grid      GridView
	Pos       Position
	Tick      uint64
	Research  ResearchView
	Tiers     TierView
	Modifiers GlobalModifiers
	Layer     Layer
	// InjectionElement is the globally selected Element all Injectors emit.
	InjectionElement Element
}

// NewTestApplyContext returns a zero-valued ApplyContext with Modifiers
// normalized and a Tiers view that returns BaseTier for every kind, so
// component tests don't have to remember the Normalized() precondition or
// construct tier scaffolding by hand. Grid and Research views remain nil —
// component tests that exercise grid/research reads must populate them
// explicitly.
func NewTestApplyContext() ApplyContext {
	return ApplyContext{
		Tiers:            newTierView(nil),
		Modifiers:        GlobalModifiers{}.Normalized(),
		InjectionElement: ElementHydrogen,
	}
}

// GridView is the read-only accessor for grid state, handed to components via
// ApplyContext. Cells are returned by value; SubjectsAt returns a freshly
// allocated slice so callers cannot mutate live grid data.
type GridView interface {
	CellAt(p Position) (Cell, bool)
	SubjectsAt(p Position) []Subject
	InBounds(p Position) bool
	Size() int
}

// ResearchView is the read-only accessor for per-Element research level.
// Absent entries return 0.
type ResearchView interface {
	Level(e Element) int
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
