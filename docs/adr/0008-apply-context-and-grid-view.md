# ADR 0008 — Apply context and read-only grid view

**Status:** accepted.
**Date:** 2026-04-24.

## Context

Every Accelerator Component today implements `sim.Component`:

```go
type Component interface {
    Kind() ComponentKind
    Apply(s Subject) (Subject, bool)
}
type Spawner interface {
    Component
    MaybeSpawn(pos Position) (Subject, bool)
}
```

`Apply` sees only the inbound Subject. It has no access to the grid, to the Subject's position on the grid (`Pos` is only implicit in the caller), to the tick number, to per-Element research levels, or to global-upgrade modifiers. `MaybeSpawn` has the Position but nothing else.

Three categories of component the game wants next are blocked by this:

1. **Neighbor-aware** (e.g. Resonator counts adjacent Resonators) — needs grid read.
2. **Research-gated** (e.g. Catalyst only triggers if research for the Subject's Element ≥ N) — needs the per-Element research level.
3. **Tick-phased or load-aware** — needs the tick counter or load cap.

Widening `Apply` also has to carry global-upgrade modifiers (ADR 0010) from GameState into components, and has to be forward-compatible with subject-emitting components (ADR 0009).

Two constraints shape the design:

- ADR 0006 froze the `(Subject, lost bool)` return shape. New capabilities must come as new **parameters** or **sibling interfaces**, never a return-shape change.
- ADR 0003 keeps `internal/sim/` headless (no ebiten imports). That doesn't block this change, but it reinforces that `ApplyContext` is a plain-data struct.

## Decision

**1. A read-only `ApplyContext` is threaded through `Apply` and `MaybeSpawn`.**

```go
// internal/sim/components.go

type Component interface {
    Kind() ComponentKind
    Apply(ctx ApplyContext, s Subject) (Subject, bool)
}
type Spawner interface {
    Component
    MaybeSpawn(ctx ApplyContext, pos Position) (Subject, bool)
}

// ApplyContext is the read-only view of world state handed to components
// during a tick. All fields are safe to read; implementations must not
// mutate any referenced object nor retain references across ticks.
type ApplyContext struct {
    Grid      GridView
    Pos       Position
    Tick      uint64
    Research  ResearchView
    Modifiers GlobalModifiers // see ADR 0010; zero-value = identity
    Layer     Layer
}
```

**2. `GridView` is an interface, not `*Grid`.**

```go
type GridView interface {
    CellAt(p Position) (Cell, bool)    // Cell returned by value
    SubjectsAt(p Position) []Subject   // fresh slice; callers can't mutate live data
    InBounds(p Position) bool
    Size() int
}
```

Handing out `*Grid` would let any component write to cells, swap `Component` pointers, or edit `Subjects`. Once one component does it, tick-ordering bugs follow and the save-replay story breaks. An interface plus an unexported `gridView` struct is ~30 lines and eliminates the entire class.

`CellAt` returns `Cell` by value. `Cell.Component` is an interface (pointer-shaped), so a caller *could* type-assert and reach in — mitigated by documentation, not structurally. The expected read pattern is inspecting `cell.Component.Kind()`, not poking fields on the concrete type.

`SubjectsAt` returns a freshly allocated slice on every call. Current subject density is low (single-digits), so the allocation cost is negligible compared to the safety win.

**3. `ResearchView` mirrors the same read-only-wrapper pattern.**

```go
type ResearchView interface {
    Level(e Element) int
}
```

Handing out `map[Element]int` would be mutable through the reference. The wrapper closes that door.

**4. `ApplyContext` is built once per per-cell visit and once per injection pass.**

`internal/sim/tick.go` builds the context at the top of each cell visit in `stepSubject`; manual injection builds the same base context for each source. The context is a small struct; a fresh value per visit keeps `Pos` and per-visit state correct without sharing instances.

`Grid` and `Research` view objects are constructed once per `Tick()` call and reused across all visits that tick. Both wrap the live `*Grid` / `map` — stale-reference bugs are not a concern because the context doesn't outlive the call frame.

**5. The five existing components update signature in lockstep.**

No adapter layer, no dual-signature phase. The set is small (Accelerator, Injector, MeshGrid, Magnetiser, Rotator) and lockstep is cheaper than deprecation. All five keep their existing behavior and ignore `ctx` in this ADR's scope — actual reads land in the ADRs and feature docs that add or rewire components.

**6. `Apply`'s return shape is unchanged.**

ADR 0006 stands. `(Subject, lost bool)` is the contract. A subject-emitting capability is designed as a sibling interface (see ADR 0009).

**7. `ApplyContext` is the carrier for future context fields.**

Anything a new component needs to read at Apply time lands on `ApplyContext` rather than extending the `Apply` signature again. Examples that are anticipated but not shipped by this ADR: `Now time.Time` for real-time effects, `Prestige PrestigeView` for Phase-4 layer state. Adding a field to `ApplyContext` is a purely additive change; adding a parameter to `Apply` is a breaking refactor across every component.

## Consequences

**Wins**
- Every component category the roadmap wants (neighbor-aware, research-gated, tick-phased, load-aware, modifier-aware) becomes expressible without touching the tick loop.
- The read-only wrapper pattern keeps "`Apply` is a pure function" a supported convention, not a hope.
- Future context fields are additive.

**Costs**
- Interface change forces all existing components and their direct tests to update signature. Small today (5 components), but anyone writing custom test helpers will need to pass a zero-valued `ApplyContext`.
- `SubjectsAt` allocates a fresh slice per call. Acceptable at current subject counts; revisit if a future component calls it in a hot loop with hundreds of on-grid subjects.
- `GridView` does not prevent pointer-shaped escape via `Cell.Component`. Documentation mitigates; structural enforcement would require copying every Component on read, which is wasteful. The boundary is "don't mutate what you read"; reviewers enforce.

## Alternatives considered

- **Pass `*Grid` and `*GameState` directly.** Rejected: maximum coupling, zero safety. One rogue component mutating grid state creates tick-ordering bugs that don't reproduce under tests.
- **Add individual parameters to `Apply` (`Apply(s, grid, pos, tick, ...)`)**. Rejected: every new context field is a breaking refactor across every component. A struct absorbs additions.
- **Make `ApplyContext` a method receiver on a separate `Evaluator` type.** Rejected: over-engineered. `Apply` is called inline from the tick loop; a receiver adds ceremony without composability gain.
- **Supply context through a package-level `Current()` function (ambient).** Rejected: hides the data dependency, breaks test isolation, and the single-goroutine tick model doesn't need it.
- **Pass only the fields each component needs via capability interfaces (`GridReader`, `ResearchReader`, etc.) on `Component` itself.** Rejected: balloons the interface surface and turns the tick loop into a capability-dispatch mesh. One `ApplyContext` is cheaper.

## Related

- `internal/sim/components.go` — new interface shapes and `ApplyContext`.
- `internal/sim/grid.go` — `gridView` impl and constructor.
- `internal/sim/state.go` — `researchView` helper.
- `internal/sim/tick.go` — per-visit context construction and threading.
- `internal/sim/components/*.go` — signature updates across all five components.
- ADR 0003 — headless sim constraint (still holds).
- ADR 0006 — return shape freeze.
- ADR 0009 — subject-emitter capability sits alongside `Apply`.
- ADR 0010 — `GlobalModifiers` rides on `ApplyContext`.
