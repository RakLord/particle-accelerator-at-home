# ADR 0018 — Laboratory upgrade tree

**Status:** accepted (Phase 4 design freeze; implementation pending).
**Date:** 2026-04-24.

> Implementation note: ADR 0019 changes Dense Packing from per-Binder capacity to Binder Store capacity via `GlobalModifiers.BinderStoreCapacityMul`.

## Context

The Laboratory tree is the third upgrade catalog plugging into the modifier pipeline (after `GlobalUpgradeCatalog` from ADR 0010 and `BondCatalog` from ADR 0016). It houses upgrades paid in Bond Points and unlocked post-prestige. See `docs/features/0023-laboratory.md` for the player-facing description and full catalog.

Three architectural questions:

1. **Storage shape.** Bonds are a set (`map[BondID]bool`) because each Bond is owned-or-not. Laboratory upgrades have *levels* (Dense Packing 0-5, Auto-Inject Speed 0-4) — set is wrong. Map of int? Slice? Per-upgrade fields on `GameState`?
2. **Modifier integration.** Laboratory upgrades have effects that are *not* purely `*GlobalModifiers` mutations: Stable Isotope copies research forward across `ResetGenesis`; Carbon Core fast-forwards the Carbon unlock chain. These touch `*GameState`, not just `*GlobalModifiers`. How should the `Apply` closure shape accommodate that?
3. **Currency reuse.** Bond Points are per-prestige meta-currency. Should `GameState.BondPoints int` be a primitive int, or wrapped in something extensible (a future "currency" type)?

## Decision

**1. `LaboratoryUpgrades map[LabUpgradeID]int` on `GameState`.**

```go
type GameState struct {
    // ...
    LaboratoryUpgrades map[LabUpgradeID]int `json:"laboratory_upgrades,omitempty"`
    BondPoints         int                  `json:"bond_points,omitempty"`
}

type LabUpgradeID string

const (
    LabCovalentMemory   LabUpgradeID = "covalent_memory"
    LabStableIsotope    LabUpgradeID = "stable_isotope"
    LabChainReaction    LabUpgradeID = "chain_reaction"
    LabCarbonCore       LabUpgradeID = "carbon_core"
    LabDensePacking     LabUpgradeID = "dense_packing"
    LabAutoInjectSpeed  LabUpgradeID = "auto_inject_speed"
)
```

`LaboratoryUpgrades[id]` is the current level, defaulting to zero (not purchased). For one-shot upgrades, level is 0 or 1. For tiered upgrades (Dense Packing, Auto-Inject Speed), level can climb to the catalog's `MaxPurchases`.

This shape mirrors `PurchasedUpgrades map[GlobalUpgradeID]bool` from ADR 0010 with one key difference — int instead of bool — to express tier level. Catalog entries report whether they are stackable; the purchase entry point enforces the cap.

**2. `LabUpgrade.Apply` takes `(*GlobalModifiers, *GameState, int level)`.**

```go
type LabUpgrade struct {
    ID           LabUpgradeID
    Name         string
    Description  string
    MaxPurchases int
    CostByLevel  []int // BP cost to advance from level i to i+1
    Apply        func(m *GlobalModifiers, s *GameState, level int)
}
```

The signature is broader than `GlobalUpgrade.Apply(*GlobalModifiers)` because Laboratory upgrades legitimately need to mutate `*GameState`:

- `Stable Isotope`: must read pre-reset research and scale-write post-reset research in `ResetGenesis`. Implemented as a "preserve research at 30%" hook called *during* `ResetGenesis`, not in `rebuildModifiers`. See point 5 below.
- `Carbon Core`: must add `BinderReserves[Carbon]` and `UnlockedElements[Carbon..]` to a fresh-Run state. Same — called from `ResetGenesis`.
- `Covalent Memory`: simpler — also a `ResetGenesis` hook. Adds `UnlockedElements[Helium] = true` after the reset wipes them.

So `Apply` has two distinct call sites:

- **From `rebuildModifiers(s)`** (per ADR 0010), passing `s.Modifiers`, `s`, `level`. Most upgrades take this path. Effects on `*GlobalModifiers` only.
- **From `ResetGenesis(s)`**, after the wipe but before `rebuildModifiers`, passing `nil, s, level`. The few upgrades that need to seed post-reset state take this path.

To keep the catalog clean, distinguish via a dispatch flag:

```go
type LabUpgrade struct {
    // ...
    AppliesIn       LabApplyPhase  // LabApplyModifiers | LabApplyResetSeed
    Apply           func(m *GlobalModifiers, s *GameState, level int)
}

type LabApplyPhase int

const (
    LabApplyModifiers   LabApplyPhase = iota  // called from rebuildModifiers
    LabApplyResetSeed                         // called from ResetGenesis post-wipe
)
```

`rebuildModifiers` only iterates Lab upgrades with `AppliesIn == LabApplyModifiers`. `ResetGenesis` iterates those with `AppliesIn == LabApplyResetSeed` after the wipe. A given upgrade is in exactly one phase — never both.

| Upgrade | Phase |
|---|---|
| Covalent Memory | LabApplyResetSeed |
| Stable Isotope | LabApplyResetSeed (snapshots research pre-wipe; see below) |
| Chain Reaction | LabApplyModifiers (modifies `InjectorRateMul`) |
| Carbon Core | LabApplyResetSeed |
| Dense Packing | LabApplyModifiers (writes `BinderStoreCapacityMul`; see ADR 0019) |
| Auto-Inject Speed I-IV | LabApplyModifiers (drives `AutoInjectCadenceTicks`) |

Note: ADR 0019 supersedes the original Dense Packing implementation detail. Dense Packing now writes `GlobalModifiers.BinderStoreCapacityMul` so future global upgrades can compose with the same Binder Store capacity field.

**3. `ResetGenesis` calls Lab upgrades with the `LabApplyResetSeed` phase.**

```go
// internal/sim/state.go (sketch — extends ADR 0014)

func ResetGenesis(s *GameState) {
    // 1. Snapshot anything Stable Isotope needs to read pre-wipe.
    var researchSnapshot map[Element]int
    if s.LaboratoryUpgrades[LabStableIsotope] > 0 {
        researchSnapshot = copyMap(s.Research)
    }

    // 2. Wipe (per ADR 0014).
    wipeGenesisFields(s)

    // 3. Apply ResetSeed Lab upgrades — they may write to s.
    for id, level := range s.LaboratoryUpgrades {
        u, ok := LabCatalog[id]
        if !ok || u.AppliesIn != LabApplyResetSeed { continue }
        u.Apply(nil, s, level)
    }

    // 4. Stable Isotope's seed step uses the snapshot.
    if researchSnapshot != nil {
        for e, v := range researchSnapshot {
            s.Research[e] = v * 30 / 100  // 30% carry
        }
    }

    // 5. Rebuild modifiers from the surviving Bonds + Lab Modifier-phase upgrades.
    rebuildModifiers(s)
}
```

Stable Isotope is a special-cased step rather than a pure `Apply` closure because it needs the pre-wipe snapshot. The closure-only design can't read across the wipe. Keeping the snapshot logic in `ResetGenesis` itself, gated by `LaboratoryUpgrades[LabStableIsotope] > 0`, is simpler than threading a "before-wipe context" through the closure signature.

**4. `rebuildModifiers(s)` is extended again.**

After ADR 0010 (global upgrades) and ADR 0016 (Bonds), add a third loop:

```go
func rebuildModifiers(s *GameState) {
    s.Modifiers = GlobalModifiers{}

    // global upgrades (ADR 0010)
    for id := range s.PurchasedUpgrades { /* ... */ }

    // bonds (ADR 0016)
    for id := range s.BondsState { /* ... */ }

    // laboratory (ADR 0018)
    for id, level := range s.LaboratoryUpgrades {
        u, ok := LabCatalog[id]
        if !ok || u.AppliesIn != LabApplyModifiers || level == 0 { continue }
        u.Apply(&s.Modifiers, s, level)
    }
}
```

Order matters when modifiers stack. Bonds and Lab upgrades that affect the same field (e.g. both touching `InjectorRateMul`) compose the same way two global upgrades do — multiplicative for `Decimal` fields, additive for `int` fields. The order of iteration is stable-but-unspecified for maps, but since the operations are commutative (multiplication is commutative), the result is deterministic regardless.

**5. `PurchaseLabUpgrade(s *GameState, id LabUpgradeID) error`.**

```go
func PurchaseLabUpgrade(s *GameState, id LabUpgradeID) error {
    u, ok := LabCatalog[id]
    if !ok { return errUnknownLabUpgrade }
    cur := s.LaboratoryUpgrades[id]
    if cur >= u.MaxPurchases { return errMaxedOut }
    if cur >= len(u.CostByLevel) { return errCatalogMisconfigured }
    cost := u.CostByLevel[cur]
    if s.BondPoints < cost { return errInsufficientBP }

    s.BondPoints -= cost
    s.LaboratoryUpgrades[id] = cur + 1
    rebuildModifiers(s)
    return nil
}
```

`CostByLevel[cur]` is the cost to advance from level `cur` to `cur+1`. Tiered upgrades (Dense Packing: `[1, 2, 3, 4, 5]`; Auto-Inject Speed: `[1, 2, 3, 4]`) lay these out explicitly. One-shot upgrades have a single-element slice.

The function rebuilds modifiers after every purchase. Cheap, predictable. No partial-state on failure (early-return on every gate).

**6. `BondPoints int` is a plain int.**

No wrapper, no currency type, no abstraction. It's a counter. If a future prestige layer introduces a second meta-currency (e.g. "Cosmic Points" from a layer above Genesis), revisit. Until then, primitive int is correct.

**7. Save schema: additive.**

`LaboratoryUpgrades`, `BondPoints` are `omitempty`. Old saves load with empty values. No envelope bump.

## Consequences

**Wins**
- Three catalogs sharing one rebuild pipeline — `GlobalUpgradeCatalog`, `BondCatalog`, `LabCatalog`.
- Tiered upgrades fit the int-map shape without special storage.
- `LabApplyPhase` cleanly separates "modifier-pipeline" upgrades from "post-reset state seeding" upgrades, without forcing every closure to handle both.
- Stable Isotope's pre-wipe snapshot is one block in `ResetGenesis`, not a buried capability on every upgrade.

**Costs**
- Three catalogs and a phase enum is more shape than the global-upgrades-only design. Justified by the genuine difference in effect type (modifiers vs. state seeding) and gating (USD vs. tokens vs. BP). Resist the urge to unify; they will diverge further as more prestige layers ship.
- `LabApplyResetSeed` upgrades are special-cased — running them at the wrong time (e.g. inside `rebuildModifiers`) would silently re-seed state on every load. Tests should assert the phase split is honored.
- Order-dependence in `rebuildModifiers` map iteration: only safe because all current modifier compositions are commutative. If a future upgrade introduces non-commutative composition (e.g. clamping after multiply), revisit and impose a stable iteration order via a sorted slice.

## Alternatives considered

- **Make Laboratory upgrades a subset of `GlobalUpgradeCatalog` with a `Currency` discriminator.** Rejected: same reason as ADR 0016's Bond/Global split — the cost surface and gating differ enough that flattening hurts more than helps.
- **Drop the `LabApplyPhase` enum; let every upgrade run in `rebuildModifiers` and have side effects there.** Rejected: would mean `Stable Isotope` re-snapshots research on every save load — running for the wrong reason. Phase separation is correct.
- **Per-upgrade fields on `GameState` instead of a map (`s.DensePackingLevel int`, `s.AutoInjectSpeedLevel int`).** Rejected: doesn't scale; new upgrades require sim changes; the map is simpler.
- **Rebuild modifiers only on action (purchase / synthesis / load), never per-tick.** This is the current design — already correct, no change needed. Mentioned for clarity.
- **Persist computed `Modifiers` and skip rebuild on Lab purchase.** Rejected per ADR 0010 — drift on retuning.

## Related

- `internal/sim/laboratory.go` (new) — `LabCatalog`, `LabUpgrade`, `PurchaseLabUpgrade`, `computeAutoInjectCadence`.
- `internal/sim/state.go` — `LaboratoryUpgrades`, `BondPoints`, `ResetGenesis` extension.
- `internal/sim/modifiers.go` — `rebuildModifiers` extended again.
- `internal/ui/laboratory_tab.go` (new) — tab UI.
- `docs/features/0023-laboratory.md` — player-facing description.
- ADR 0002 — additive save fields.
- ADR 0010 — first catalog; modifier pipeline being extended.
- ADR 0014 — `ResetGenesis` calls `LabApplyResetSeed` upgrades; `LaboratoryUpgrades` is preserved.
- ADR 0015 — Dense Packing's effect is read by `EffectiveBinderCapacity`.
- ADR 0016 — second catalog; Bonds plug into the same pipeline.
- ADR 0017 — Auto-Inject Speed I-IV drive `AutoInjectCadenceTicks`.
