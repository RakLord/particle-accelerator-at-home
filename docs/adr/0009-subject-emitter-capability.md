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

**2. The tick loop dispatches via a type assertion; extras are admitted after the advance-subjects filter.**

`stepSubject` type-asserts for `Splitter` and, when matched, appends extras to a per-tick `pending []Subject` slice owned by `advanceSubjects`. `advanceSubjects` runs its in-place filter over live Subjects first, then admits `pending` into `g.Subjects` under the `EffectiveMaxLoad` cap.

```go
// internal/sim/tick.go (sketch; concrete lines will shift)

// advanceSubjects:
var pending []Subject
alive := g.Subjects[:0]
for _, sub := range g.Subjects {
    collected, lost := s.stepSubject(&sub, &pending)
    // ... existing collect/lost/keep-alive handling ...
}
g.Subjects = alive
cap := s.EffectiveMaxLoad()
for _, e := range pending {
    if s.CurrentLoad+e.Load > cap { continue }  // per-extra MaxLoad gate
    g.Subjects = append(g.Subjects, e)
    s.CurrentLoad += e.Load
}

// stepSubject (inner loop, per-cell-visit):
if sp, ok := cell.Component.(Splitter); ok {
    self, extras, destroyed := sp.ApplySplit(ctx, *sub)
    *sub = self
    *pending = append(*pending, extras...)
    if destroyed { return false, true }
} else {
    *sub, destroyed = cell.Component.Apply(ctx, *sub)
    if destroyed { return false, true }
}
```

Why defer admission? Appending to `g.Subjects` from inside `stepSubject` races with the in-place filter (`alive := g.Subjects[:0]`) — the filter's final `g.Subjects = alive` assignment silently drops anything appended mid-loop. A separate `pending` slice keeps extras safe until the filter is done.

Side effect of the order: the input Subject's `Load` is freed (via the `lost` branch) before extras are admitted. A full grid with a MaxLoad equal to the input's Load can admit one extra — matching the "partial admission" policy.

**3. Extras are load-accounted individually and may be partially admitted.**

Each extra is a full Subject and costs its own Load. `MaxLoad` is enforced per-extra, not per-ApplySplit call. If the grid has no room for all extras, the ones that fit are admitted and the rest are silently dropped — same semantics as manual Injector admission.

Alternative: reject the whole ApplySplit atomically if any extra can't fit. Rejected because it couples a single component's emit outcome to unrelated grid pressure; partial admission matches Injector behavior and is easier to reason about.

**4. Extras bypass the cell they were emitted from.**

Extras start positioned at the emitter's cell with `InDirection` set by the Splitter itself and `StepProgress` at the standard spawn value (`StepProgressPerCell / 2`, matching Injector). They are not re-fed through the emitter's `Apply` on the same tick — that would loop. They travel on the next tick like any other Subject.

This matches how a real T-junction behaves: both output pipes are different paths, not another trip through the junction.

**5. Duplicator is the first and only Splitter in this wave.**

`internal/sim/components/duplicator.go` implements `Splitter`. Its `ApplySplit` consumes the incoming Subject (`lost=true`) and emits two extras — one going the Subject's incoming perpendicular `.Left()` direction, one going `.Right()`. Both carry the incoming Subject's `Element`, `Mass`, `Speed`, `Magnetism`, `Load`, halved and rounded per its tier stats (see ADR 0011). Gameplay detail lives in `docs/features/component-duplicator.md`.

**6. `Splitter` is not a sprite or UI concept.**

`Kind()` and the render layer treat Splitters the same as any other component. The capability is a runtime dispatch detail.

## Consequences

**Wins**
- Components that don't emit extras remain completely unchanged. `Apply` stays a `(Subject, bool)` function.
- The pattern is forward-compatible: future "absorber" components that consume *neighboring* subjects (not just incoming) can land as another sibling interface without further reshape.
- MaxLoad enforcement is consistent with manual Injector admission — one mental model for subject admission.

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
