# ADR 0015 — Binder component and per-Element reserves

**Status:** accepted (Phase 4 design freeze; implementation pending).
**Date:** 2026-04-24.

## Context

The Binder is the new prestige-feeder Component (see `docs/features/component-binder.md`). It is a typed endpoint that accepts Subjects of a specific Element and stores them in a reserve, sacrificing all $USD and research that Subject would have generated.

Three architectural questions need answering:

1. **Where does the reserve live?** Per-cell on the grid, or per-Element on `GameState`?
2. **How is the Element binding stored?** On the Component itself (like Pipe's `Orientation`), or as a separate per-cell config?
3. **How does "destroy on full" express in the existing Component interface?** `Apply` already supports the `lost=true` return — does it need any extension?

The first question matters for capacity calculation. If reserves are per-cell, "Hydrogen reserve full" is a per-Binder concept; if per-Element, it's an aggregate. The proposal in `tmp.md` reads as per-Element (`BinderCapacity map[Element]int`), but is silent on whether two Binders pool or split.

The second question matters for consistency. Pipe and Rotator carry `Orientation` as a runtime-mutable Component field (ADR 0006). Element binding has the same "set on placement, never changes after" lifecycle.

The third question is a verification, not an extension: `Component.Apply` returning `(Subject{}, true)` already means "destroy this Subject". The question is just whether `Apply` has access to the state it needs.

## Decision

**1. Reserves are stored per-Element on `GameState`. Multiple Binders of the same Element pool.**

```go
type GameState struct {
    // ...
    BinderReserves map[Element]int `json:"binder_reserves,omitempty"`
}
```

A Hydrogen Binder banking a Subject increments `BinderReserves[ElementHydrogen]`. A second Hydrogen Binder elsewhere on the grid increments the same counter. The capacity check is:

```go
EffectiveBinderCapacity(s, e) = countBindersOf(s.Grid, e) * BinderBaseCapacity[e] * densePackingMultiplier(s)
```

Pooling makes "place more Binders" a meaningful capacity strategy. It also means a single Crystallise call drains from one shared pool, regardless of which physical Binder absorbed which Subject — much simpler UX than asking the player to pick which Binder to crystallise from.

The alternative — per-cell reserves — was rejected. It implies UX where the player crystallises one Binder at a time, has to balance fill levels across Binders, and is told "you can't crystallise; this Binder isn't full yet". For an idle game, this is friction without a corresponding strategic upside.

**2. Element binding is a Component field, not a separate per-cell map.**

```go
// internal/sim/components/binder.go

type Binder struct {
    Element  sim.Element
    // (no other state — reserve lives on GameState)
}

func (b *Binder) Kind() sim.ComponentKind { return sim.KindBinder }
```

Same shape as Pipe (`Orientation`). Set at placement via the placement flow's Element picker. Never reassigned without removing and re-placing.

The picker is a small modification to the placement UI: when placing a `KindBinder` Component, show an Element selector (default Hydrogen, all unlocked Elements selectable). The player's choice flows into the Component constructor.

**3. `Component.Apply` needs `*GameState` write access — it goes through `ApplyContext`.**

The Binder's `Apply` must:

- Increment `BinderReserves[Binder.Element]` if the incoming Subject's Element matches and the reserve is below capacity.
- Destroy the Subject in either case (match-and-bank, or no-match).

Today's `ApplyContext` (ADR 0008) carries a *read-only* `Grid` view and immutable modifier reads. Binder needs to mutate one specific field on `GameState`. Two options:

- **Add `*sim.GameState` to `ApplyContext`.** Breaks ADR 0008's read-only contract for *every* component.
- **Add a narrow side-effect channel — `Banker` capability — analogous to `Splitter` in ADR 0009.**

Choose the second.

```go
// internal/sim/components.go

// Banker is an optional capability for components that bank incoming
// Subjects into a per-Element reserve on GameState. Components that
// don't bank should implement Component, not Banker.
type Banker interface {
    Component
    // ApplyBank is the bank-aware variant of Apply. It returns whether
    // the incoming Subject was banked (true → reserve incremented),
    // along with the standard (Subject, lost) tuple. The tick loop
    // increments the reserve when banked == true and respects the
    // standard lost return for grid removal.
    ApplyBank(ctx ApplyContext, s Subject) (out Subject, lost bool, banked bool, element Element)
}
```

The tick loop handles the increment so the Component itself never writes to `GameState`:

```go
// internal/sim/tick.go (sketch)

if bk, ok := cell.Component.(Banker); ok {
    out, lost, banked, e := bk.ApplyBank(ctx, *sub)
    if banked {
        if state.BinderReserves[e] < state.EffectiveBinderCapacity(e) {
            state.BinderReserves[e]++
        }
        // overfull case: banked == true was reported by the component
        // assuming it could fit; the cap check here is the final guard.
        // The Subject is still destroyed (lost==true) regardless.
    }
    *sub = out
    if lost { return false, true }
} else if sp, ok := cell.Component.(Splitter); ok {
    // ... existing Splitter dispatch
} else {
    // ... existing plain Apply dispatch
}
```

The capacity check is **double-gated**: the component reports its intent (`banked=true` means "I would like this Subject banked"), the tick loop confirms there is room. If full, the Subject is still destroyed (the `lost=true` came from the component) but no reserve increment happens. This matches the gameplay rule: full Binder still destroys.

`Banker` is sibling to `Splitter`; the dispatch order is `Banker` → `Splitter` → plain. A single Component implements at most one of these; any future Component that needs to be both banks AND emits is a separate ADR.

**4. The full-Binder notification is sim-driven via the helper-notifications system.**

When the tick loop's capacity check fails (Subject was banked-intent but reserve was full), it emits a notification once per session per Element:

```go
if banked && state.BinderReserves[e] >= state.EffectiveBinderCapacity(e) {
    notifyBinderFullOnce(state, e)
}
```

`notifyBinderFullOnce` uses a transient session-only set on `GameState` (not saved) to ensure one notification per Element per session. See `docs/features/helper-notifications.md`.

**5. Wrong-Element entry is destroyed silently.**

A Subject of the wrong Element entering a Binder is destroyed, no banking, no notification. The Component's `ApplyBank` returns `(Subject{}, lost=true, banked=false, element=anything)` for wrong-Element cases. Same outcome as wrong-axis Pipe entry — consistent player mental model.

**6. Per-Binder capacity is a const map; Dense Packing multiplies through `EffectiveBinderCapacity`.**

```go
// internal/sim/components/binder.go

var BinderBaseCapacity = map[sim.Element]int{
    sim.ElementHydrogen: 15,
    sim.ElementHelium:    8,
    sim.ElementLithium:  30,
    sim.ElementCarbon:  100,
    // heavier elements TBD
}
```

`GameState.EffectiveBinderCapacity(e Element) int` reads the base, multiplies by `2^DensePackingLevel` (capped at 5 → ×32), and multiplies by the count of Binders of that Element on the grid.

```go
func (s *GameState) EffectiveBinderCapacity(e Element) int {
    base := BinderBaseCapacity[e]
    count := s.Grid.CountComponents(KindBinder, func(c Component) bool {
        return c.(*components.Binder).Element == e
    })
    multiplier := 1 << min(s.LaboratoryUpgrades[LabDensePacking], 5)
    return base * count * multiplier
}
```

A new helper `Grid.CountComponents(kind, filter)` lives on the grid type — a simple linear scan, fine at MVP grid sizes.

**7. Save shape: per-Element reserve only.**

```go
BinderReserves map[Element]int `json:"binder_reserves,omitempty"`
```

The Binder Component's `Element` field is part of its serialized Component data (the existing Component-serialization path handles per-Component fields like `Orientation`). No additional per-cell save state.

## Consequences

**Wins**
- Reserves are a single map; capacity is a one-line computation.
- Component never writes to `GameState`; the tick loop is the only mutator. Preserves ADR 0008's contract for everyone except the dispatching tick loop.
- `Banker` is a clean, narrow capability — same precedent as `Splitter` (ADR 0009).
- "Place more Binders" is a real strategic lever for capacity.

**Costs**
- Tick loop now has a third dispatch branch (Banker, Splitter, plain). Three branches is still tractable; if a fourth lands, refactor into a registry.
- `Grid.CountComponents` is a linear scan called inside `EffectiveBinderCapacity`. At MVP grid size (5×5 → ~100 cells worst case incl. growth) this is cheap, but if `EffectiveBinderCapacity` is called per-tick from inside a hot path, cache the result. Capacity changes only on placement/removal/Lab upgrade — not per tick — so caching is straightforward when needed.
- The Element picker UX is new surface — placement flow gets a per-Component config step for Binder. Mitigated by reusing the Codex Element selector pattern.
- Wrong-Element silent destruction can confuse a new player who misroutes. The collection log (see `docs/features/collection-log.md`) shows the loss as a normal lost-Subject event, which is a soft cue. Stronger feedback (e.g. a one-time tutorial notification) is out of scope for MVP.

## Alternatives considered

- **Per-cell reserves.** Rejected: more UX friction (per-Binder crystallise) without strategic upside, and pool size is more legible to players.
- **Pass `*GameState` into `ApplyContext`.** Rejected: breaks ADR 0008 for every Component; opens the door to scope creep.
- **Make Binder a `Spawner` variant that emits "no Subject".** Rejected: misuses the existing capability; obscures intent.
- **Store Element binding in a `map[GridCell]Element` on `GameState`.** Rejected: loosens Component encapsulation; per-Component fields exist for a reason.
- **Cap Binder capacity globally instead of per-Element-per-Binder.** Rejected at gameplay layer — the per-Element capacity table is a load-bearing knob.

## Related

- `internal/sim/components.go` — `Banker` interface.
- `internal/sim/components/binder.go` — implementation.
- `internal/sim/kinds.go` — `KindBinder` constant.
- `internal/sim/tick.go` — Banker dispatch, capacity check, notification trigger.
- `internal/sim/state.go` — `BinderReserves`, `EffectiveBinderCapacity`.
- `internal/sim/grid.go` — `CountComponents` helper.
- `docs/features/component-binder.md` — player-facing description.
- ADR 0006 — orientation/per-Component config precedent.
- ADR 0008 — `ApplyContext` read-only contract (preserved).
- ADR 0009 — `Splitter` capability precedent.
- ADR 0014 — `BinderReserves` is wiped by `ResetGenesis`.
- ADR 0016 — what reserves are spent on (Token crystallisation).
