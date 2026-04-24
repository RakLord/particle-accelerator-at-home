# ADR 0010 — Global modifier pipeline

**Status:** accepted (Phase 2 pipeline shipped; Phase 6 purchase flow pending).
**Date:** 2026-04-24.

## Context

`docs/overview.md` names **global upgrades** ("all Collectors +10%", "Inject cooldown 2× faster", "all subjects start at Speed +1") as one of four progression axes. Nothing in the current code supports them:

- `collectValue` (`internal/sim/economy.go:64-73`) computes collection value inline with no hook.
- Manual injection cooldown reads its base cooldown directly.
- `SimpleAccelerator.Apply` reads its own `SpeedBonus` directly.
- `MaxLoad` is a fixed starter value on `GameState` with no upgrade path.

Two questions bound the design:

1. Where does the aggregated multiplier live? Scattered across the code (each feature reads its own upgrade set) or centralized into one struct that every hot path reads?
2. What is the source of truth? The aggregated multipliers themselves, or the underlying set of purchased upgrades?

Scattered reads couple every component to the upgrade catalog and make the hot paths harder to reason about. Centralizing into a single struct lets every read site consult one object.

Persisting aggregated multipliers and persisting purchased upgrades both work, but only one avoids drift when upgrade definitions change: if we persist multipliers and later retune an upgrade's effect, existing saves keep the old number forever. Persisting the set of purchased upgrades and **deriving** multipliers on load means every retuning applies on next load.

## Decision

**1. `GlobalModifiers` aggregates every active global effect into one struct on `GameState`.**

```go
// internal/sim/modifiers.go (new)

// GlobalModifiers aggregates player-purchased global upgrades. Components and
// hot paths read fields directly; they are never mutated from gameplay code —
// rebuildModifiers(s) is the only writer. Decimal fields have multiplicative
// semantics (zero-value means "no bonus" after Normalized()). Integer fields
// are additive.
type GlobalModifiers struct {
    CollectorValueMul       bignum.Decimal
    InjectorRateMul         bignum.Decimal
    AcceleratorSpeedBonus   int
    MagnetiserBonusMul      bignum.Decimal
    ResearchPerCollectBonus int
    MaxLoadBonus            int
}

// Normalized returns a copy with zero-valued Decimal fields promoted to 1 so
// downstream multiplication is safe. Called by tick.go when building
// ApplyContext.
func (m GlobalModifiers) Normalized() GlobalModifiers
```

The Decimal-vs-int split is deliberate: `AcceleratorSpeedBonus` stays integer because Speed is integer in the simulation. `ResearchPerCollectBonus` stays integer because research is an int counter. Everything else is multiplicative and uses `bignum.Decimal` to compose over long play sessions (ADR 0004).

**2. `GameState.Modifiers` is derived from `GameState.PurchasedUpgrades`.**

```go
type GameState struct {
    // ... existing fields
    PurchasedUpgrades map[GlobalUpgradeID]bool `json:"purchased_upgrades,omitempty"`
    Modifiers         GlobalModifiers          `json:"modifiers,omitempty"`
}
```

`Modifiers` is the fast-read cache. `PurchasedUpgrades` is the source of truth. Every purchase and every save-load runs:

```go
func rebuildModifiers(s *GameState) {
    s.Modifiers = GlobalModifiers{} // zero
    for id := range s.PurchasedUpgrades {
        up, ok := GlobalUpgradeCatalog[id]
        if !ok { continue }          // unknown id = retired upgrade; ignore
        up.Apply(&s.Modifiers)
    }
}
```

`Modifiers` is persisted too, but only as a cache — `rebuildModifiers` is called on every load after unmarshal, so the persisted value is immediately overwritten from the authoritative purchase set. Persisting it anyway keeps saves self-describing when inspected, and the `omitempty` tag means freshly created saves don't carry a redundant zero value.

**3. Modifier read sites are enumerated and load-bearing.**

| Field | Read site |
|---|---|
| `CollectorValueMul` | `collectValue(s, research, mods)` in `economy.go` — multiplies the final value |
| `InjectorRateMul` | `GameState.EffectiveInjectionCooldownTicks()` — divides effective manual injection cooldown, minimum 1 tick |
| `AcceleratorSpeedBonus` | `SimpleAccelerator.Apply` — flat add on top of the component's own `SpeedBonus` (pre-ADR-0011) or tier bonus (post-ADR-0011) |
| `MagnetiserBonusMul` | `Magnetiser.Apply` — multiplies the component's per-apply bonus |
| `ResearchPerCollectBonus` | `advanceSubjects` in `tick.go` — added to `s.Research[Element]` on collection |
| `MaxLoadBonus` | `GameState.EffectiveMaxLoad()` — additive on top of base `MaxLoad`, read by `injectorSpawns` and UI |

`EffectiveMaxLoad()` is a new read-only helper so base `MaxLoad` stays an unambiguous "un-upgraded cap". Upgrade sources (prestige, events, globals) feed the helper, not the base field.

**4. Upgrade definitions carry `Apply(*GlobalModifiers)` closures, not static values.**

```go
// internal/sim/global_upgrades.go (new)

type GlobalUpgradeID string

type GlobalUpgrade struct {
    ID               GlobalUpgradeID
    Name             string
    Description      string
    Cost             bignum.Decimal
    RequiresElement  Element
    RequiresResearch int
    Apply            func(m *GlobalModifiers)
}

var GlobalUpgradeCatalog = map[GlobalUpgradeID]GlobalUpgrade{
    "collector_value_10": {
        ID: "collector_value_10",
        Name: "Collector Coils I",
        Description: "All Collectors yield 10% more $USD.",
        Cost: bignum.MustParse("1000"),
        RequiresElement: ElementHydrogen,
        RequiresResearch: 5,
        Apply: func(m *GlobalModifiers) {
            m.CollectorValueMul = m.CollectorValueMul.Mul(bignum.MustParse("1.10"))
        },
    },
    // ... more
}
```

Closures instead of raw multipliers let a single upgrade touch multiple fields ("all subjects start faster AND load is higher"). Every upgrade is responsible for being idempotent under `rebuildModifiers` — the same purchase set must always produce the same `GlobalModifiers`.

**5. Purchase is a single entry point with gating.**

```go
func PurchaseGlobalUpgrade(s *GameState, id GlobalUpgradeID) error
```

Checks: upgrade exists, not already purchased, research/element prerequisite met, `USD >= Cost`. On success: deduct USD, set `PurchasedUpgrades[id] = true`, call `rebuildModifiers(s)`. All-or-nothing; no partial state on failure.

**6. Save-compat is additive.**

Both new fields use `omitempty`. An old save with neither field unmarshals with an empty `PurchasedUpgrades` map and a zero-valued `Modifiers` struct. Post-unmarshal `rebuildModifiers(s)` produces the same zero modifiers, and `Normalized()` promotes zero-valued Decimals to identity multipliers. Old players load into a game with no global upgrades yet — exactly the state they left.

No save envelope bump required (ADR 0005 already established that additive fields don't require one).

**7. `ApplyContext` carries `Modifiers` by value.**

`GlobalModifiers` is small and copy-cheap. Components read `ctx.Modifiers.X` directly. Normalization happens once per tick, not once per Apply call — the tick loop calls `Normalized()` on `s.Modifiers` and stashes the result in the per-tick context.

## Consequences

**Wins**
- Global upgrades become a sealed extension surface: new upgrades are catalog entries, zero core code changes.
- Retuning an upgrade's effect applies retroactively on next load — no migration needed for balance changes.
- `Modifiers` is the one cache; every hot path reads exactly the field it needs.

**Costs**
- Two fields on `GameState` for the same concept (purchased set + derived cache). `rebuildModifiers` must be called in exactly the right places (save-load + after each purchase). Missing a call leaves `Modifiers` stale.
- Closures in `GlobalUpgradeCatalog` aren't serializable — the catalog is code, not data. Changing an upgrade's effect requires a code deploy; adding a new upgrade requires a code deploy. Acceptable for an incremental game with all-client state.
- `MaxLoad` now has a base vs. effective split. Existing code that reads `s.MaxLoad` directly must be audited and switched to `s.EffectiveMaxLoad()` where the upgrade should apply.

## Alternatives considered

- **Persist `Modifiers` as the source of truth; no `PurchasedUpgrades`.** Rejected: balance retuning doesn't propagate to existing saves. Players locked into old values forever.
- **Scatter upgrade reads across the code (each feature consults `PurchasedUpgrades` itself).** Rejected: every hot path pays a map lookup and the Element/research threshold logic, instead of reading one struct field.
- **Use a trait-like interface per modifier category (`CollectorModifier`, `InjectorModifier`, ...).** Rejected: over-structured. The struct is small and likely to stay small; six fields don't justify six interfaces.
- **Make upgrades stackable via a purchase count instead of a bool.** Rejected for this wave: the first four upgrades are explicitly one-shot unlocks. Stackable variants can land later by swapping `map[ID]bool` for `map[ID]int` without disturbing the pipeline.

## Related

- `internal/sim/modifiers.go` (new) — `GlobalModifiers` and `Normalized`.
- `internal/sim/global_upgrades.go` (new) — catalog, `PurchaseGlobalUpgrade`, `rebuildModifiers`.
- `internal/sim/state.go` — `Modifiers`, `PurchasedUpgrades`, `EffectiveMaxLoad`.
- `internal/sim/economy.go` — `collectValue` signature change.
- `internal/sim/tick.go` — `Normalized()` called per tick, threaded via `ApplyContext`.
- `internal/sim/save.go` — call `rebuildModifiers(s)` after unmarshal.
- `docs/features/global-upgrades.md` — player-facing description.
- ADR 0002 — versioned save schema (additive change, no bump).
- ADR 0004 — bignum core (backs the multiplier fields).
- ADR 0005 — established that additive save fields are allowed.
- ADR 0008 — `ApplyContext` carries `Modifiers`.
