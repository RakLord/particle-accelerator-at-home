# ADR 0009 — Subject-emitter capability

**Status:** accepted.
**Date:** 2026-04-24.

## Context

The Duplicator (T-junction) component needs to transform one incoming Subject into two outgoing Subjects travelling in different directions. `Component.Apply` returns exactly one Subject (or destroys it, per ADR 0006), so the existing interface cannot express this.

Three shapes were considered for accommodating subject-count changes:

1. Change `Apply` to return `([]Subject, bool)`. Every existing component returns a one-element slice. Uniform but noisy — every component allocates or returns a slice header.
2. Let components mutate `*Grid.Subjects` directly from inside `Apply` via the `ApplyContext`. Violates the read-only contract in ADR 0008.
3. Add a **sibling capability interface** that components opt into when they need to emit extras. The tick loop checks for the capability and dispatches appropriately.

Option 3 matches the pattern already established by `Spawner`: a narrow, opt-in interface for a specific capability, checked via a type assertion in the tick loop. `Apply`'s contract stays identical for the 99% of components that don't emit.

## Decision

**1. A new `Splitter` capability interface sits alongside `Apply`.**

```go
// internal/sim/components.go

// Splitter is an optional capability for components that, on Apply, may
// produce extra Subjects in addition to transforming the incoming one.
// Components that don't emit extras should implement Component, not Splitter.
type Splitter interface {
    Component
    // ApplySplit is the emitter-aware variant of Apply. The first return
    // is the transformed incoming Subject (same shape as Apply). Extras
    // are additional Subjects the component emits this cell visit; each
    // is appended to the grid with its Load charged against MaxLoad
    // individually. If the incoming Subject is lost, extras are still
    // emitted — a Splitter can consume the input and emit replacements.
    ApplySplit(ctx ApplyContext, s Subject) (self Subject, extras []Subject, lost bool)
}
```

A component implements **either** `Component` alone **or** `Splitter` (which embeds `Component`). The tick loop prefers `Splitter` when present.

**2. The tick loop dispatches via a type assertion.**

```go
// internal/sim/tick.go (sketch; concrete lines will shift)
if sp, ok := cell.Component.(Splitter); ok {
    self, extras, lost := sp.ApplySplit(ctx, *sub)
    *sub = self
    if lost {
        // incoming is destroyed; see note (3) on extras
    }
    for _, e := range extras {
        if s.CurrentLoad+e.Load > s.MaxLoad { continue } // MaxLoad gate applies per extra
        g.Subjects = append(g.Subjects, e)
        s.CurrentLoad += e.Load
    }
} else {
    *sub, lost := cell.Component.Apply(ctx, *sub)
    // existing flow
}
```

The `Splitter` branch runs in `stepSubject` after the position and direction have been updated for the incoming Subject, same sequencing as the current `Apply` call.

**3. Extras are load-accounted individually and may be partially admitted.**

Each extra is a full Subject and costs its own Load. `MaxLoad` is enforced per-extra, not per-ApplySplit call. If the grid has no room for all extras, the ones that fit are admitted and the rest are silently dropped — same semantics as the Injector spawn gate.

Alternative: reject the whole ApplySplit atomically if any extra can't fit. Rejected because it couples a single component's emit outcome to unrelated grid pressure; partial admission matches Injector behavior and is easier to reason about.

**4. Extras bypass the cell they were emitted from.**

Extras start positioned at the emitter's cell with `InDirection` set by the Splitter itself and `StepProgress` at the standard spawn value (`SpeedDivisor / 2`, matching Injector). They are not re-fed through the emitter's `Apply` on the same tick — that would loop. They travel on the next tick like any other Subject.

This matches how a real T-junction behaves: both output pipes are different paths, not another trip through the junction.

**5. Duplicator is the first and only Splitter in this wave.**

`internal/sim/components/duplicator.go` implements `Splitter`. Its `ApplySplit` consumes the incoming Subject (`lost=true`) and emits two extras — one going the Subject's incoming perpendicular `.Left()` direction, one going `.Right()`. Both carry the incoming Subject's `Element`, `Mass`, `Speed`, `Magnetism`, `Load`, halved and rounded per its tier stats (see ADR 0011). Gameplay detail lives in `docs/features/component-duplicator.md`.

**6. `Splitter` is not a sprite or UI concept.**

`Kind()` and the render layer treat Splitters the same as any other component. The capability is a runtime dispatch detail.

## Consequences

**Wins**
- Components that don't emit extras remain completely unchanged. `Apply` stays a `(Subject, bool)` function.
- The pattern is forward-compatible: future "absorber" components that consume *neighboring* subjects (not just incoming) can land as another sibling interface without further reshape.
- MaxLoad enforcement is consistent with Injector spawns — one mental model for subject admission.

**Costs**
- Tick loop has two dispatch branches instead of one. Small cost, centralized.
- Silent partial admission of extras can surprise players ("why did only one path fire?"). Mitigated by keeping extras within MaxLoad in practice; revisit if it becomes a confusion vector.
- `ApplySplit` signature is slightly asymmetric vs. `Apply` (extra return value). The naming difference is deliberate — a misread `Apply` call still compiles; a misread `ApplySplit` does not.

## Alternatives considered

- **Single `Apply` returning `[]Subject`.** Rejected: every component pays slice allocation for a one-element result. Noise on the 99% case to serve the 1%.
- **Mutate `*Grid` from `Apply` via `ApplyContext`.** Rejected: contradicts ADR 0008's read-only contract. Opens the door to every other mutation rationale.
- **Return an "event queue" from `Apply` (`[]Event`)**. Rejected: the Subject-lifecycle events we have today (`transformed`, `destroyed`) are already expressible in the return tuple. An event queue is premature for one new capability.
- **Chain multiple `Apply` calls against the same component for each extra.** Rejected: requires mutable component state to track "which of my outputs am I producing right now" — fragile and breaks purity.

## Related

- `internal/sim/components.go` — `Splitter` interface.
- `internal/sim/tick.go` — dispatch branch in `stepSubject`.
- `internal/sim/components/duplicator.go` — first implementor.
- `docs/features/component-duplicator.md` — gameplay description.
- ADR 0006 — Apply return shape freeze (respected; Splitter is a sibling).
- ADR 0008 — `ApplyContext` is the shared input.
