# ADR 0017 — Auto-injection

**Status:** accepted (Phase 4 design freeze; implementation pending).
**Date:** 2026-04-24.

## Context

Auto-injection is the prestige layer's idle-mode injection mechanism (see `docs/features/auto-injection.md`). It is unlocked by the Benzene Bond and tunable through the Laboratory's Auto-Inject Speed I-IV upgrades.

`docs/features/manual-injection.md` already mentions this in passing: "Auto-injection is intentionally not enabled yet, but can later call the same `GameState.Inject` path." That's the foundation. The remaining decisions:

1. **Where does the auto-fire timer live?** On `GameState`? On `GlobalModifiers`? Elsewhere?
2. **How do auto-cadence upgrades and `InjectorRateMul` (manual cooldown speed) compose?** Stack? Or are they independent?
3. **How is the unlock expressed — a boolean modifier, a flag derived from `BondsState`, or some other gate?**

The composition question is load-bearing. If auto-cadence and `InjectorRateMul` stack multiplicatively, Speed IV (cadence ×0.4 → 4s) plus Acetylene (cooldown ×0.75) plus Chain Reaction (×0.5 effective rate) compresses the auto-fire interval below tick granularity, and the simulation breaks.

## Decision

**1. Auto-cadence and `InjectorRateMul` are independent. They do not stack.**

`InjectorRateMul` shortens the **manual cooldown** — the time between consecutive `Inject()` admissions, regardless of who calls it. Auto-cadence is the **scheduling interval** — how often the auto path *attempts* to call `Inject()`.

The two compose as: every `AutoInjectCadenceTicks`, the auto path calls `Inject()`; that call is admitted only if `InjectionCooldown` is ready. If the cadence is shorter than the effective cooldown, the auto path will fail-fast on most ticks (no admission, no injection) — wasted scheduling but no simulation break.

In practice, the Lab Speed upgrades cap the cadence at 4s (Speed IV). The base cooldown is 5s. With Acetylene, effective cooldown is 5s × 0.75 = 3.75s. So Speed IV + Acetylene means the auto path schedules every 4s, but each attempt is admitted because the cooldown (3.75s) has already cleared. Auto fires every 4s effectively. **No simulation break, no requirement to stack.**

If a future upgrade pushes effective cooldown below 3s, the cadence stays the bottleneck. If a future Lab upgrade pushes cadence below cooldown, the cooldown becomes the bottleneck. Either way, the slower of the two governs.

**2. Two new fields on `GlobalModifiers`.**

```go
type GlobalModifiers struct {
    // ... existing fields
    AutoInjectEnabled    bool  `json:"auto_inject_enabled,omitempty"`
    AutoInjectCadenceTicks int `json:"auto_inject_cadence_ticks,omitempty"`
}
```

`AutoInjectEnabled` is set by Benzene's `Apply` closure. `AutoInjectCadenceTicks` is set by the Laboratory tree's Auto-Inject Speed upgrades — the highest-purchased level wins, computed from a small lookup table.

```go
// internal/sim/laboratory.go (sketch)

var autoInjectCadenceByLevel = []int{
    100, // level 0 (Benzene unlock baseline) = 10s @ 10 Hz
     80, // level 1 = 8s
     60, // level 2 = 6s
     50, // level 3 = 5s
     40, // level 4 = 4s
}

func computeAutoInjectCadence(s *GameState) int {
    if !s.BondsState[BondBenzene] {
        return 0  // unset — auto-inject not enabled
    }
    level := s.LaboratoryUpgrades[LabAutoInjectSpeed]
    if level >= len(autoInjectCadenceByLevel) {
        level = len(autoInjectCadenceByLevel) - 1
    }
    return autoInjectCadenceByLevel[level]
}
```

`rebuildModifiers(s)` writes `AutoInjectCadenceTicks` from this helper. Note that it reads `BondsState` directly — `AutoInjectCadenceTicks` only has meaning if `AutoInjectEnabled` is true, but maintaining that invariant inside `rebuildModifiers` is simple.

**3. The auto path is a new branch in the tick loop's injection step, after the manual cooldown advance.**

```go
// internal/sim/tick.go (sketch)

// existing: advance manual injection cooldown
if state.InjectionCooldown > 0 {
    state.InjectionCooldown--
}

// new: advance auto-inject scheduler
if state.Modifiers.AutoInjectEnabled && state.AutoInjectActive {
    state.AutoInjectTickCounter++
    if state.AutoInjectTickCounter >= state.Modifiers.AutoInjectCadenceTicks {
        state.AutoInjectTickCounter = 0
        _ = state.Inject()  // best-effort; admission may fail
    }
}
```

`state.Inject()` is the existing entry point used by the manual button. Reusing it means auto and manual fire through identical admission, identical cooldown management, identical Max Load checking. No second code path to maintain.

**4. `AutoInjectActive bool` on `GameState` is the persistent UI toggle.**

```go
type GameState struct {
    // ...
    AutoInjectActive      bool `json:"auto_inject_active,omitempty"`
    AutoInjectTickCounter int  `json:"-"` // transient
}
```

`AutoInjectActive` persists across saves and across prestige (ADR 0014). On Benzene unlock, the toggle defaults to `false` — the player must opt in. They will see a new "Auto Inject" button next to manual Inject; turning it on engages the cadence loop.

`AutoInjectTickCounter` is `json:"-"` (transient): a saved game resumes with a fresh cycle, not a partial one. This is deliberate. Persisting the counter means a player who saves at counter `99/100` and reloads gets one immediate fire on resume, which is confusing. Transient counter means "next fire is `AutoInjectCadenceTicks` from resume", which matches player expectation.

**5. The Auto-Inject toggle UI is a small button next to the existing Inject button.**

Visible only when `state.Modifiers.AutoInjectEnabled` is true (i.e., Benzene synthesised). Two states: `Auto: ON` (highlighted, tied to `AutoInjectActive == true`), `Auto: OFF` (default). Clicking toggles `AutoInjectActive`. No confirmation modal — toggling has no destructive consequence.

The button can stay visible across runs even when no Injectors are placed — the cadence loop is a no-op without Injectors (the auto path's `state.Inject()` returns early with no work). No special-case suppression needed.

**6. No concept of "auto-inject queue" or "burst auto-inject".**

The auto path fires the same single-call action as the manual button: each placed Injector emits one Subject (subject to admission). There is no batching, no holdover, no "I missed the last cycle, fire twice next time" logic. If admission fails (cooldown not ready, Max Load saturated), that cycle is lost.

Players who want higher throughput buy `Chain Reaction` (Lab) or stack more Injectors. Auto-inject is a scheduler, not a multiplier.

**7. Save compatibility.**

`AutoInjectActive` is `omitempty` and defaults to `false` — old saves load with auto off. `AutoInjectTickCounter` is `json:"-"` and not serialized at all. The two new `GlobalModifiers` fields are derived by `rebuildModifiers(s)` post-load, so they are always consistent with `BondsState` on load. No save-envelope bump.

## Consequences

**Wins**
- Reuses the manual injection path entirely — one admission code path, one cooldown rule.
- Independent composition with `InjectorRateMul` avoids the cadence-vs-cooldown stacking trap.
- Cadence upgrades plug into the same `rebuildModifiers` pipeline (ADR 0010, extended by ADR 0016 and 0018) — no separate upgrade-application surface.
- Transient `AutoInjectTickCounter` keeps save semantics predictable.

**Costs**
- Two `GlobalModifiers` fields are tightly coupled (`Enabled` is meaningless without `CadenceTicks`, and vice versa). A future maintainer might be tempted to merge into `AutoInjectCadenceTicks int` with `0 == disabled`. Rejected here because it muddles "is this feature unlocked" with "how fast is it" — the bool is the unlock semantic, the int is the tuning. Keep both.
- The cadence-vs-cooldown independence means the player can buy upgrades that don't help (Speed IV with no Acetylene means cadence is the bottleneck; Acetylene without Speed IV means cooldown ≤ cadence on most levels). Some wasted purchases. Mitigated by the catalog ordering — Speed I-IV's BP costs are clearly tied to the Lab tree, separate from manual upgrades.
- Each auto-inject attempt that fails admission still runs through `Inject()`'s cooldown / Max Load checks. Cheap, but a tight loop. At 10 Hz tick rate and 1 attempt per `AutoInjectCadenceTicks` ticks, this is at most one extra check per ~40-100 ticks. Negligible.

## Alternatives considered

- **Stack `InjectorRateMul` into the auto cadence (multiplicative).** Rejected: breaks the simulation when stacked deep upgrades push cadence below 1 tick. Independent composition is safer and easier to reason about.
- **Single `AutoInjectCadenceTicks` field with `0 == disabled`.** Rejected: blurs unlock vs. tuning. Two fields are more honest about intent.
- **Persist `AutoInjectTickCounter`.** Rejected: confusing reload behavior; the transient counter matches "next fire is cadence ticks from now" which players expect.
- **Auto-inject as a Component (placed on grid like an Injector variant).** Rejected: would need its own placement, cost, sprite, balance. Auto is a scheduler-level feature, not a component-level one. The Lab tree is the right surface.
- **`GameState.AutoInject Cooldown` mirroring `InjectionCooldown`, separate from auto-cadence.** Rejected: doubles the cooldown bookkeeping for no clear gameplay difference. Cadence is enough.

## Related

- `internal/sim/injection.go` — auto path in tick loop, `Inject()` reuse.
- `internal/sim/modifiers.go` — `AutoInjectEnabled`, `AutoInjectCadenceTicks` fields.
- `internal/sim/laboratory.go` — `computeAutoInjectCadence`, Speed I-IV upgrades.
- `internal/sim/bonds.go` — Benzene's `Apply` sets `AutoInjectEnabled`.
- `internal/sim/state.go` — `AutoInjectActive`, `AutoInjectTickCounter`.
- `internal/ui/inject_panel.go` — toggle UI.
- `docs/features/auto-injection.md` — player-facing description.
- `docs/features/manual-injection.md` — referenced "auto-injection later" call-out.
- ADR 0010 — modifier pipeline extended with `AutoInjectEnabled` / `Cadence` fields.
- ADR 0014 — `AutoInjectActive` persists, `AutoInjectTickCounter` resets.
- ADR 0016 — Benzene Bond's `Apply` closure that flips `AutoInjectEnabled`.
- ADR 0018 — Lab Speed I-IV that drive `AutoInjectCadenceTicks`.
