# ADR 0014 — Carbon prestige layer (reset model and persistence)

**Status:** accepted (Phase 4 design freeze; implementation pending).
**Date:** 2026-04-24.

## Context

`docs/overview.md` lists Phase 4 as "Reset layer / Max Load upgrades / Grid-size upgrades" but commits no specifics. `internal/sim/layer.go` defines `Layer` with `LayerGenesis` as scaffolding only — no transition logic, no soft-reset function, no meta-currency.

`internal/sim/state.go:117` ships `HardReset()` for the Settings menu's destructive "wipe save" affordance. That is not a prestige reset: it returns `GameState` to its zero value with no carry-over.

The prestige layer needs:

1. A trigger gate — when does the Prestige button appear?
2. A reset function — what does prestige wipe, what does it preserve?
3. A persistence model — which fields survive, how do they flow into the new run?
4. An integration with the modifier pipeline so persistent bonuses are immediately active post-reset.

The trigger gate is settled by gameplay design (see `docs/features/prestige-genesis-ascension.md`): synthesise one Bond, gate opens. The reset scope is settled by gameplay design too (the **hardest** scope: research and unlocks both wiped). The remaining decisions are how to express that in code without breaking save-compat or doubling the reset surface.

## Decision

**1. `ResetGenesis(s *GameState)` is a sibling to `HardReset`, not a refactor of it.**

```go
// internal/sim/state.go

// ResetGenesis returns the game to a fresh-Run state while preserving
// prestige-layer progression. Called when the player presses the Prestige
// button. Distinct from HardReset (which wipes everything including
// prestige progression).
func ResetGenesis(s *GameState) {
    // wipe Genesis-layer state
    s.USD = startingUSD
    s.Grid = NewGrid(s.GridSize)
    s.Owned = startingInventory()
    s.Research = make(map[Element]int)
    s.UnlockedElements = map[Element]bool{ElementHydrogen: true}
    s.BinderReserves = make(map[Element]int)
    s.TokenInventory = make(map[Element]int)
    s.CurrentLoad = 0
    s.AutoInjectTickCounter = 0

    // increment run counter for UI / future achievements
    s.RunCount++

    // BondsState, BondPoints, LaboratoryUpgrades, BestStats, AutoInjectActive,
    // and Layer are NOT touched.

    // Rebuild the modifier cache from the surviving sources so Bond effects
    // are active for the new run.
    rebuildModifiers(s)
}
```

`HardReset` stays as-is for the Settings "wipe save" path. The two functions co-exist, distinguishable by name and call site. **Do not** parameterize `HardReset` with a "preserve prestige" flag — keeping them separate keeps each one's intent legible at the call site.

**2. The list of preserved fields is enumerated, not derived.**

Preserved across `ResetGenesis`:

- `BondsState map[BondID]bool` — synthesised compounds.
- `BondPoints int` — unspent Bond Points.
- `LaboratoryUpgrades map[LabUpgradeID]int` — purchased Laboratory upgrades and their levels.
- `BestStats` — lifetime achievements (already exists, already not reset by anything but `HardReset`).
- `AutoInjectActive bool` — UI-only toggle preference.
- `Layer Layer` — stays `LayerGenesis` (later layers are out of scope).
- `RunCount int` — incremented, not preserved as-is.

Anything not in this list is reset. Future fields default to "reset" unless explicitly listed here. The function is checklist-driven on purpose: forgetting a new field is a *bug* but the bug is "reset accidentally wiped my prestige progression" — visible the first time a tester prestiges, not silently corrupted.

**3. The Prestige trigger is `len(BondsState) >= 1`, evaluated by the UI, not by sim.**

```go
// internal/ui/prestige_button.go

func (p *PrestigeButton) Visible(s *sim.GameState) bool {
    return len(s.BondsState) >= 1
}
```

Sim does not need a `CanPrestige(s)` helper — the condition is one map-length check, and putting it behind a function would imply gating logic that doesn't exist (no $USD threshold, no run-time floor). If future gating gets richer, introduce the helper then.

**4. `RunCount int` is a new, save-persisted field.**

Used by the UI ("Run #5") and future achievement / pacing telemetry. Increments inside `ResetGenesis`. Has no gameplay effect on its own — it is informational. `omitempty` so old saves load as zero.

**5. New `GameState` fields, all `omitempty`.**

```go
type GameState struct {
    // ... existing fields

    // Prestige-layer state
    BinderReserves     map[Element]int             `json:"binder_reserves,omitempty"`
    TokenInventory     map[Element]int             `json:"token_inventory,omitempty"`
    BondsState         map[BondID]bool             `json:"bonds_state,omitempty"`
    BondPoints         int                         `json:"bond_points,omitempty"`
    LaboratoryUpgrades map[LabUpgradeID]int        `json:"laboratory_upgrades,omitempty"`
    AutoInjectActive   bool                        `json:"auto_inject_active,omitempty"`
    AutoInjectTickCounter int                      `json:"-"` // transient
    RunCount           int                         `json:"run_count,omitempty"`
}
```

`AutoInjectTickCounter` is `json:"-"` (transient): a paused/saved game should resume with a fresh cadence cycle, not continue an interrupted one. The next auto-fire happens after `AutoInjectCadenceTicks` from load.

**6. Save schema is additive — no envelope bump.**

ADR 0002 establishes that additive fields don't require a save version bump; ADR 0010 reaffirmed it for `PurchasedUpgrades`/`Modifiers`. The seven new fields above are all additive. Old saves load with zero values for each — equivalent to "never prestiged, never banked, never bonded" — which is the intended pre-prestige state.

**7. `rebuildModifiers(s)` is called by `ResetGenesis` after the wipe.**

ADR 0010 requires `rebuildModifiers` after every state load and every modifier-source mutation. `ResetGenesis` mutates `BondsState`'s effective surface (no — it doesn't, but it mutates surrounding fields and the *post-reset* modifier set should still reflect the surviving Bonds). Calling it post-wipe ensures `state.Modifiers` is consistent for the new run's first tick.

The rebuild is cheap (one pass over a small map). The cost is amortized across an action the player does every several minutes. No optimization needed.

**8. `HardReset` is updated to also clear the new prestige fields.**

```go
func (s *GameState) HardReset() {
    *s = NewGameState()  // existing contract — full zero
}
```

`NewGameState()` returns the zero starting state. The new prestige fields are zero-initialized there too (empty maps / zero ints), so `HardReset` automatically wipes them. No code change needed in `HardReset` itself, but a unit test should assert post-`HardReset` state has `BondsState == nil` etc., to catch any future drift.

## Consequences

**Wins**
- The reset surface is two named functions with non-overlapping intent.
- Prestige progression is a checklist-driven preservation set — visible, auditable, easy to extend.
- Save schema is fully additive; no migration; old saves work unmodified.
- `rebuildModifiers` reuse means Bond effects are active on the first tick of the new run with zero special-case code.

**Costs**
- Two reset functions to keep in sync as new fields land. Mitigated by a unit test that asserts `ResetGenesis` preserves exactly the listed fields (and only those).
- `RunCount` is a new field with no current consumer beyond UI display. Carries a small "what's this for?" cost for new readers; the comment block on `ResetGenesis` is the documentation.
- `AutoInjectTickCounter`'s transient-on-save behavior is a quiet edge case; comment-marked, but a future maintainer who serializes it inadvertently will subtly desync auto-inject across save/load.

## Alternatives considered

- **Parameterize `HardReset(preservePrestige bool)`.** Rejected: the call sites have different intent, and a boolean flag obscures the difference. Two named functions are clearer.
- **Persist a `Layer Layer` value that's already advanced past `LayerGenesis`.** Rejected for this wave: there are no downstream layers yet. `Layer` stays `LayerGenesis` and prestige is a *within-Genesis* reset. When later layers ship, they will introduce a separate `AscendLayer()` operation.
- **Award Bond Points per-prestige (1 per run) instead of per-Bond.** Rejected at the gameplay-design layer (see review notes in plan file). Per-Bond, complexity-weighted is the chosen model.
- **Drive the Prestige UI from a `CanPrestige(s)` sim helper.** Rejected for now: the condition is trivial. Will reconsider if multiple gates accrete.
- **Reset *only* the grid and inventory, preserving research/unlocks.** Rejected at the gameplay-design layer (the hardest reset is the chosen model).

## Related

- `internal/sim/state.go` — `ResetGenesis`, new fields, `RunCount`.
- `internal/sim/layer.go` — existing scaffold; not changed by this ADR.
- `internal/sim/save.go` — `rebuildModifiers(s)` already runs post-load, no change.
- `docs/features/prestige-genesis-ascension.md` — player-facing description.
- ADR 0002 — additive save fields, no version bump.
- ADR 0010 — modifier pipeline that Bonds and Laboratory upgrades extend.
- ADR 0015 — Binder component (sim-side reserves source).
- ADR 0016 — Token and Bond economy (state shapes preserved by `ResetGenesis`).
- ADR 0017 — Auto-injection (state shapes including the transient counter).
- ADR 0018 — Laboratory upgrade tree (state shape preserved).
