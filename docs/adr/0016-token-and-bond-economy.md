# ADR 0016 — Token and Bond economy

**Status:** accepted (Phase 4 design freeze; implementation pending).
**Date:** 2026-04-24.

## Context

Tokens (per-Element crystallisations) and Bonds (synthesised compounds) are the two prestige-currency layers between *banked Subjects* and *permanent modifiers* (see `docs/features/bonds-and-tokens.md`).

Three architectural questions:

1. **Where do Tokens and Bonds live on `GameState`?** As maps? Slices? Bitsets?
2. **How do Bond effects integrate with the existing modifier pipeline (ADR 0010)?** Do they share `PurchasedUpgrades` and `GlobalUpgradeCatalog`, live in their own catalog, or some hybrid?
3. **How is the Token cost function shaped?** Hard-coded sequence, formula, table, or other?

The first question is mechanical but matters for save shape. The second is the real design decision: ADR 0010 set up `rebuildModifiers(s)` to derive `Modifiers` from `PurchasedUpgrades`. Bonds need to plug into that — a Bond's effect is, structurally, identical to a `GlobalUpgrade`'s effect (a closure that mutates `*GlobalModifiers`). The question is whether to *merge* the two catalogs or run them in parallel.

The third question has a clean answer (formula over table) but the formula details affect how reorderable the catalog is.

## Decision

**1. New `GameState` fields, all maps, all `omitempty`.**

```go
type GameState struct {
    // ...
    TokenInventory map[Element]int     `json:"token_inventory,omitempty"`
    BondsState     map[BondID]bool     `json:"bonds_state,omitempty"`
    // BinderReserves and BondPoints live here too — see ADR 0014/0015/0018.
}

type BondID string
```

`TokenInventory[e]` is the count of Tokens of Element `e` the player owns and has not yet spent. `BondsState[id]` is `true` iff the Bond is synthesised. Tokens are wiped on prestige; Bonds persist (ADR 0014).

`BondID` is a string newtype to keep it lexically distinct from `GlobalUpgradeID` and `LabUpgradeID`. Catalog keys live alongside their data, like `GlobalUpgradeCatalog`:

```go
const (
    BondMethane   BondID = "methane"
    BondAcetylene BondID = "acetylene"
    BondEthylene  BondID = "ethylene"
    BondBenzene   BondID = "benzene"
    BondDiamond   BondID = "diamond"
)
```

**2. Bonds run as a parallel catalog to `GlobalUpgradeCatalog`, plugging into the same `rebuildModifiers` pipeline.**

ADR 0010's `rebuildModifiers(s)` reads `PurchasedUpgrades` and applies catalog closures. Extend it to also read `BondsState` and apply the Bond catalog's closures:

```go
// internal/sim/modifiers.go (extended)

func rebuildModifiers(s *GameState) {
    s.Modifiers = GlobalModifiers{} // zero

    for id := range s.PurchasedUpgrades {
        up, ok := GlobalUpgradeCatalog[id]
        if !ok { continue }
        up.Apply(&s.Modifiers)
    }

    for id := range s.BondsState {
        b, ok := BondCatalog[id]
        if !ok { continue }
        b.Apply(&s.Modifiers)
    }

    // Laboratory upgrades extend this further — see ADR 0018.
}
```

The Bond catalog mirrors the global-upgrade catalog shape:

```go
type Bond struct {
    ID          BondID
    Name        string
    Formula     string
    Description string
    TokenCost   map[Element]int   // e.g. Methane: {Carbon: 1, Hydrogen: 4}
    BondPoints  int               // BP awarded on first synthesis
    Apply       func(m *GlobalModifiers)
}

var BondCatalog = map[BondID]Bond{
    BondMethane: {
        ID:          BondMethane,
        Name:        "Methane",
        Formula:     "CH₄",
        Description: "The simplest hydrocarbon.",
        TokenCost:   map[Element]int{ElementCarbon: 1, ElementHydrogen: 4},
        BondPoints:  1,
        Apply: func(m *GlobalModifiers) {
            m.CollectorValueMul = m.CollectorValueMul.Mul(bignum.MustParse("1.15"))
        },
    },
    // ... others
}
```

**Why parallel catalogs, not merged?** A Bond is purchased in *Tokens* with a per-recipe cost shape; a global upgrade is purchased in *$USD* with a single-number cost. Their gating predicates are different. Their UI surfaces are different (Bonds tab vs. shop tab). Merging into a single `Upgrade` type would require an `UpgradeKind` discriminator and conditional cost validation logic — all to flatten two coherent abstractions into one. Parallel catalogs keep each surface clean and the rebuild pipeline still treats them uniformly.

**Why share `rebuildModifiers`?** Because the *output* (the `GlobalModifiers` struct) is the same. A Bond's `Apply` closure is signature-identical to a `GlobalUpgrade`'s. The pipeline's job is "turn a set of unlocks into a modifier struct"; both kinds of unlock fit that job.

**3. `CrystallisationCost` is a formula, not a table.**

```go
// internal/sim/economy.go

const tokenCostBase = 10
const tokenCostGrowth = 5

func CrystallisationCost(e Element, owned int) int {
    cost := tokenCostBase
    for i := 0; i < owned; i++ {
        cost *= tokenCostGrowth
    }
    return cost
}
```

So `CrystallisationCost(e, 0) == 10`, `(e, 1) == 50`, `(e, 2) == 300` (with the 6× factor between 50 and 300 as stated in the design — match the table by adjusting the growth factor or hard-code an early irregularity if needed; the design doc says "~5× each step", which `tokenCostGrowth = 5` matches with growth factor 6 from 50→300 being a documented sequence quirk).

**Note on the published sequence (`10, 50, 300, 1500, 8000`):** the increments are 5×, 6×, 5×, ~5.3×. If the design intends an exact table, replace the formula with a `var tokenCostTable = []int{10, 50, 300, 1500, 8000, 40000, 200000}` lookup with a fallback `(table[len-1] * 5) ^ (n - len + 1)` extrapolation for n beyond the table. The senior dev should pick one of these on implementation; the spec is ambiguous.

The cost is per-Element. The function takes `Element` for forward extensibility (a future per-Element cost-multiplier upgrade can read it) but doesn't currently use it.

**4. `CrystalliseToken(s, e)` is the only place Tokens are minted.**

```go
func CrystalliseToken(s *GameState, e Element) error {
    cost := CrystallisationCost(e, s.TokenInventory[e])
    if s.BinderReserves[e] < cost {
        return errInsufficientReserve
    }
    s.BinderReserves[e] -= cost
    s.TokenInventory[e]++
    return nil
}
```

No partial state on failure. No automatic crystallisation from the tick loop — the player must press the Crystallise button per Element.

**5. `SynthesiseBond(s, id)` is the only place Bonds are minted, and the only place BP is awarded.**

```go
func SynthesiseBond(s *GameState, id BondID) error {
    b, ok := BondCatalog[id]
    if !ok { return errUnknownBond }
    if s.BondsState[id] { return errAlreadySynthesised }

    for e, n := range b.TokenCost {
        if s.TokenInventory[e] < n {
            return errInsufficientTokens
        }
    }

    for e, n := range b.TokenCost {
        s.TokenInventory[e] -= n
    }
    s.BondsState[id] = true
    s.BondPoints += b.BondPoints

    rebuildModifiers(s)
    return nil
}
```

Two-pass on the Token check: first verify all Element costs are satisfiable, then deduct. Avoids a partial-deduct on a multi-Element recipe where the second Element comes up short.

`rebuildModifiers(s)` is called after every successful synthesis. That is the only state mutation that requires it from this surface.

**6. Bond effects can stack with global upgrades, multiplicatively.**

Bond `Apply` closures multiply into `GlobalModifiers` fields the same way global upgrades do. Methane's `CollectorValueMul ×1.15` and a future `Collector Coils II ×1.10` global upgrade compose to `1.265×`. This falls out of ADR 0010's design — Bonds don't introduce a new composition rule.

**7. Save schema is additive.**

`TokenInventory`, `BondsState`, and `BondPoints` are all `omitempty`. Old saves load with empty values. `rebuildModifiers(s)` runs post-load (existing behavior) and produces zero modifiers for an empty `BondsState` — same gameplay as pre-prestige.

If a Bond is removed from `BondCatalog` in a future release while an existing save has it owned, the missing-id branch in `rebuildModifiers` skips it — the same retired-upgrade tolerance as global upgrades. Player-visible: the Bond's bonus silently vanishes. Acceptable for an incremental game in active development.

## Consequences

**Wins**
- Bonds and global upgrades share the modifier output shape — one `GlobalModifiers` struct, one rebuild pipeline.
- Bond catalog is a sealed extension surface: adding a new Bond is a catalog entry, no core code changes.
- Token cost is a one-function formula; tuning is a constant change.
- All-or-nothing synthesis avoids partial-state bugs on multi-Element recipes.

**Costs**
- Two parallel catalogs (`GlobalUpgradeCatalog` + `BondCatalog`) — and a third coming with Laboratory (ADR 0018). Pattern is consistent but the senior dev should resist the urge to refactor into one shared type until the third catalog has its own constraints visible.
- `CrystallisationCost` formula vs. table ambiguity: the proposal's published sequence has a 6× step that pure-5× growth doesn't reproduce. Senior dev decision required at implementation time. Recommend table for first ship (matches spec exactly), formula later when scaling beyond the table.
- `BondsState map[BondID]bool` is a set-shaped map; using a `[]BondID` would be smaller in JSON. Map keeps lookups O(1) and serializes adequately at this scale.

## Alternatives considered

- **Merge `BondCatalog` into `GlobalUpgradeCatalog` with a `Kind` discriminator.** Rejected: their cost surfaces and gating are different enough that the merge introduces conditional logic everywhere a "cost" is evaluated. Parallel catalogs are clearer.
- **Compute Bond Points from formula complexity (e.g. `sum(TokenCost values)`).** Rejected at gameplay layer: BP is hand-tuned per-Bond by the catalog author for narrative balance, not derived. Formula derivation hides design intent in a function.
- **Auto-crystallise when reserve hits the next threshold.** Rejected: removes a meaningful player decision, and means a player who *wants* to bank past Token N to unlock something cheaper would have to disable an automatic feature. Manual is the better default.
- **Token cost as a per-Element-tuned table (different growth per Element).** Rejected: every Element having the same cost curve means recipes-that-require-N-of-an-Element are the only difficulty knob. Per-Element curves would multiply the tuning surface for marginal gain.
- **Persist computed `Modifiers` rather than rerun `rebuildModifiers` on synthesis.** Rejected per ADR 0010 — drift on retuning, tested precedent.

## Related

- `internal/sim/bonds.go` (new) — `BondCatalog`, `Bond`, `SynthesiseBond`.
- `internal/sim/economy.go` — `CrystallisationCost`, `CrystalliseToken`.
- `internal/sim/modifiers.go` — `rebuildModifiers` extended to read `BondsState`.
- `internal/sim/state.go` — `TokenInventory`, `BondsState`, `BondPoints`.
- `internal/ui/bonds_tab.go` (new) — list view, Crystallise + Synthesise.
- `docs/features/bonds-and-tokens.md` — player-facing description.
- ADR 0002 — additive save fields.
- ADR 0010 — modifier pipeline being extended.
- ADR 0014 — what `ResetGenesis` preserves vs. wipes.
- ADR 0015 — Binder reserves are the input to crystallisation.
- ADR 0017 — Benzene Bond's effect (auto-injection unlock).
- ADR 0018 — Laboratory upgrade tree, third catalog plugging into the same pipeline.
