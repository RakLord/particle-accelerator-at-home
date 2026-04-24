# ADR 0011 — Component tier primitive

**Status:** accepted.
**Date:** 2026-04-24.

## Context

`docs/overview.md` lists **per-Component tiers** (e.g. Simple Accelerator T1 → T3, `+1` → `+3` Speed) as one of four progression axes. The existing code has no home for this:

- `ComponentCatalog` (`internal/sim/component_cost.go`) is keyed by `ComponentKind` with no tier dimension.
- `SimpleAccelerator.SpeedBonus` is an integer field initialized to `1` in `PlaceFromTool` and never upgraded.
- `Magnetiser.Bonus` and `Injector.SpawnInterval` are in the same shape — placement defaults, no progression.

Three shapes were considered:

1. **Mint new `ComponentKind`s per tier** (`"accelerator_t1"`, `"accelerator_t2"`, `"accelerator_t3"`). Explodes the registry and the cost catalog. Each tier becomes a distinct purchasable from the player's view, with separate Owned counters — awkward because "upgrading" is really "selling T1 and buying T2".
2. **Per-instance tier field.** Each placed component stores its own `TierLevel`. Upgrades target individual instances. Adds per-instance save state, introduces "which of my 10 Accelerators is upgraded?" UI complexity.
3. **Global tier per component kind.** `GameState.ComponentTiers[KindAccelerator] = 2` means every Accelerator on the board (and every future one) is T2. Upgrades are global, one-shot-per-tier, like research.

Overview.md phrases tiers as attributes of the component itself ("Simple Accelerator T1 → T3"), not of a specific placed instance. Option 3 matches that language, is cheapest to implement, and sidesteps per-instance UI. Per-instance tiers can still be introduced later without reshaping this pipeline.

## Decision

**1. Tier is a per-kind integer on `GameState`, defaulting to 1.**

```go
// internal/sim/tier.go (new)

type Tier int
const BaseTier Tier = 1

// TierView is the read-only accessor passed to components via ApplyContext.
type TierView interface {
    For(kind ComponentKind) Tier
}
```

```go
// internal/sim/state.go (addition)
type GameState struct {
    // ... existing fields
    ComponentTiers map[ComponentKind]Tier `json:"component_tiers,omitempty"`
}
```

Absent key → `BaseTier`. The `tierView` wrapper (unexported, similar pattern to `gridView` in ADR 0008) implements `TierView.For(kind)` with the absent-key default baked in.

**2. `ApplyContext` gains a `Tiers TierView` field.**

```go
type ApplyContext struct {
    Grid      GridView
    Pos       Position
    Tick      uint64
    Research  ResearchView
    Tiers     TierView        // NEW
    Modifiers GlobalModifiers
    Layer     Layer
}
```

Tierable components read `ctx.Tiers.For(MyKind)` inside `Apply`. Non-tierable components ignore the field. No interface change to `Component` itself — tier-awareness is a usage pattern, not a capability.

**3. Per-kind stat tables live next to each tierable component.**

Stat tables are flat int (or Decimal) slices indexed by tier. Index-out-of-range clamps to the highest defined entry, which makes "we introduced T4 in the shop but forgot to update the table" loud in tests.

```go
// internal/sim/components/accelerator.go (sketch)

var acceleratorSpeedByTier = []int{0, 1, 2, 3} // idx=0 unused; T1=+1, T2=+2, T3=+3

func (a *SimpleAccelerator) Apply(ctx sim.ApplyContext, s sim.Subject) (sim.Subject, bool) {
    // ... directional entry check unchanged ...
    tier := int(ctx.Tiers.For(sim.KindAccelerator))
    if tier < int(sim.BaseTier) { tier = int(sim.BaseTier) }
    if tier >= len(acceleratorSpeedByTier) { tier = len(acceleratorSpeedByTier) - 1 }
    s.Speed += acceleratorSpeedByTier[tier] + ctx.Modifiers.AcceleratorSpeedBonus
    return s, false
}
```

`SimpleAccelerator.SpeedBonus` field (currently a hardcoded `1` set in `PlaceFromTool`) is **removed**. Tier drives the stat; no per-instance override. Same pattern applies to `Magnetiser.Bonus` (removed in favor of a tier table) and `MeshGrid` divisor (tier table if we tier it).

**4. Tierable kinds for this wave: Accelerator, Magnetiser, Mesh Grid.**

| Kind | Tier 1 | Tier 2 | Tier 3 |
|---|---|---|---|
| Accelerator | `+1` Speed | `+2` | `+3` |
| Magnetiser | `+1` Magnetism | `+2` | `+3` |
| Mesh Grid | `÷2` Speed | `÷3` | `÷4` |

Tuning is illustrative; final values may differ at implementation time.

**Non-tierable (this wave):** Injector (per-instance config — direction, interval; Element is selected globally in the Codex), Rotator (per-instance orientation), Collector (not a Component). Duplicator and Catalyst are tierable in principle but ship at T1 only to keep the first drop small; their stat tables exist with a single entry so later tiers are additive.

**5. Tier upgrades are their own catalog, distinct from global upgrades.**

```go
// internal/sim/tier.go

type TierUpgradeInfo struct {
    Tier             Tier
    Cost             bignum.Decimal
    RequiresElement  Element
    RequiresResearch int
}

// TierUpgradeCatalog is the shop-side data. Entries are ordered by Tier.
// Purchasing T3 requires T2 to already be the current tier for that kind.
var TierUpgradeCatalog = map[ComponentKind][]TierUpgradeInfo{
    KindAccelerator: {
        {Tier: 2, Cost: bignum.MustParse("500"), RequiresElement: ElementHydrogen, RequiresResearch: 3},
        {Tier: 3, Cost: bignum.MustParse("5000"), RequiresElement: ElementHydrogen, RequiresResearch: 15},
    },
    // ... etc
}

func NextTierUpgrade(s *GameState, kind ComponentKind) (TierUpgradeInfo, bool)
func PurchaseTierUpgrade(s *GameState, kind ComponentKind) error
```

`PurchaseTierUpgrade` advances `ComponentTiers[kind]` by exactly one, deducting `NextTierUpgrade`'s cost. Prerequisite: research/element gate met, USD sufficient, a next tier exists.

Kept distinct from `GlobalUpgradeCatalog` (ADR 0010) because:
- The data shape is different (ordered list per kind vs. flat map of unlocks).
- The UI affordance is different (one "upgrade" button per kind vs. a list of named upgrades).
- A single reader can find "all tier progressions" in one catalog without grep-matching across closure bodies.

**6. Cost curve for *buying more of a component* stays single-tier for now.**

`ComponentCatalog` (`internal/sim/component_cost.go`) still has one purchase curve per kind (`Base`, `Growth`, and optional soft-cap fields). Buying your 10th Accelerator uses the same curve regardless of the global tier. Tier upgrades and purchase cost are orthogonal pricing dimensions.

If this becomes wrong (e.g. T3 Accelerators should cost more to buy than T1), the catalog can grow a per-tier override without disturbing the tier primitive itself.

**7. Save-compat is additive.**

`ComponentTiers` uses `omitempty`. An old save without the field loads with an empty map; `For(kind)` returns `BaseTier` for every kind, producing identical pre-ADR behavior. No envelope bump.

The removal of `SimpleAccelerator.SpeedBonus` and `Magnetiser.Bonus` fields is a **save-shape change**. It's backward-compatible via `UnmarshalJSON` nil-default (already the pattern `MeshGrid` uses). Old saves with `{"speed_bonus": 1}` unmarshal into the new struct harmlessly (JSON ignores unknown fields by default in Go with `encoding/json`). The removed field is silently dropped; tier behavior takes over.

## Consequences

**Wins**
- Tier progression lands as one new map on `GameState` plus one catalog. Minimal touch surface.
- Every existing placed component immediately "upgrades" when the player buys the next tier — no per-instance rework.
- Non-tierable components stay non-tierable; interface surface unchanged.
- Stat tables are simple Go slices; retuning is a one-line edit.

**Costs**
- Tier upgrades feel "free" for existing placed components, which is a deliberate gameplay choice but could cheapen the expense of buying lots of T1s in bulk. Balance via research-gate prerequisites and cost values.
- Per-instance differentiation is not possible (every Accelerator on the grid is the same tier). If a future gameplay idea needs per-instance tiers, this ADR has to be revisited — but the migration path (add a `Tier` field to the component struct, default from global) is additive.
- Two purchase surfaces (Components, Tier Upgrades) next to Global Upgrades (ADR 0010). Three shop tabs risks clutter; see `docs/features/component-tiers.md` for UI grouping.

## Alternatives considered

- **Mint kinds per tier (`"accelerator_t2"`).** Rejected — explodes registry, awkward "sell T1 / buy T2" semantics.
- **Per-instance `TierLevel` field.** Rejected — adds per-instance save state and per-instance upgrade UI for no gameplay win in this wave. Forward-compat path is still open.
- **Unify tier upgrades into `GlobalUpgradeCatalog`.** Rejected — the data shape is list-per-kind, not flat-map-of-unlocks. `rebuildModifiers` order-independence would require `max`-semantics on integer fields to stay correct, which is a footgun on a subsystem that should be order-insensitive by construction.
- **Store tier as a modifier on `GlobalModifiers` (`AcceleratorTier int`).** Rejected — mixes progression-level with multiplier-like cross-cutting upgrades. Two concepts, two homes.
- **Compute tier stats from a formula (`bonus = tier`)** instead of a table. Rejected — non-linear retunings (e.g. T3 Accelerator is `+3` then T4 jumps to `+5`) are easier in a table than in a formula. Tables also surface the full progression on a single screen for designers.

## Related

- `internal/sim/tier.go` (new) — `Tier`, `TierView`, `TierUpgradeCatalog`, purchase entry point.
- `internal/sim/state.go` — `ComponentTiers` field.
- `internal/sim/tick.go` — construct `tierView` and thread via `ApplyContext`.
- `internal/sim/components/accelerator.go` — drop `SpeedBonus`, read tier.
- `internal/sim/components/magnetiser.go` — drop `Bonus`, read tier.
- `internal/sim/components/mesh_grid.go` — read tier for divisor.
- `internal/input/input.go` — `PlaceFromTool` no longer sets per-instance tier stats.
- `docs/features/component-tiers.md` — player-facing progression.
- ADR 0002 — additive save rule.
- ADR 0005 — component purchase cost remains orthogonal to tiers.
- ADR 0008 — `ApplyContext` carries `Tiers`.
- ADR 0010 — global upgrades stay separate.
