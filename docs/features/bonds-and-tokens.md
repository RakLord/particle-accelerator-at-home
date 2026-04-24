# Bonds and Tokens

**Status:** Phase 4.

## Concept

The prestige loop has two intermediate currencies between *banked Subjects* and *permanent bonuses*:

- **Tokens** are per-Element crystallisations of banked Subjects. Owned per Element on `GameState.TokenInventory[Element]`.
- **Bonds** are synthesised compounds. Each Bond is a recipe of Tokens spent in exchange for a permanent global modifier and a small amount of Bond Points. Owned as a set of compound IDs on `GameState.BondsState`.

The flow on a Carbon-era run:

```
Inject Subjects → route to Binder → reserve fills →
crystallise Token (consumes reserve) → spend Tokens → synthesise Bond →
modifier active for all future runs + 1-3 Bond Points awarded.
```

Tokens are wiped on prestige. Bonds and the Bond Points they award persist.

## Tokens — per-Element crystallisation

Every Element has its own Token type. The cost to crystallise the *next* Token is a function of how many of that Element's Tokens the player already owns:

| Tokens already owned (this Element) | Subjects required to crystallise next |
|---:|---:|
| 0 | 10 |
| 1 | 50 |
| 2 | 300 |
| 3 | 1 500 |
| 4 | 8 000 |
| n ≥ 5 | next ≈ 5× previous |

The cost is **per-Element**: owning 2 Carbon Tokens does not affect the cost of the first Hydrogen Token. Each Element has its own scaling wall.

The cost is read from `BinderReserves[Element]` only — the player chooses to crystallise via the Bonds tab UI, and the function is:

```go
func CrystalliseToken(s *GameState, e Element) error
//   precondition:  BinderReserves[e] >= CrystallisationCost(e, TokenInventory[e])
//   on success:    BinderReserves[e] -= cost; TokenInventory[e] += 1
```

Crystallisation is manual. Subjects in the Binder reserve do not auto-convert.

### Cost intent

On a first Carbon run the player should realistically afford 1-2 Tokens *per Element* at most. Per-Element scaling means compounds that require multiple Tokens of the same Element (Benzene 6C+6H, Diamond 12C) are deliberately multi-run goals, not first-run achievements. Heavier-Element runs are expected to bring Token-gain bonuses (planned, not in MVP) that flatten this curve.

## Bonds — compound catalog

A Bond is a compound the player has synthesised. Synthesis spends Tokens and is permanent — once owned, a Bond never leaves `BondsState`. Each Bond:

- Applies a permanent modifier to `GameState.Modifiers` via the same closure-based pipeline as global upgrades (ADR 0010).
- Awards a one-time amount of Bond Points (see `docs/features/laboratory.md`).
- Marks its row in the Bonds tab as owned.

### MVP catalog (Carbon + Hydrogen only)

| Bond | Formula | Token cost | Modifier effect | Bond Points awarded |
|---|---|---|---|---:|
| Methane | CH₄ | 1C + 4H | `CollectorValueMul ×1.15` | 1 |
| Acetylene | C₂H₂ | 2C + 2H | `InjectorRateMul ×0.75` (cooldown 25% faster) | 1 |
| Ethylene | C₂H₄ | 2C + 4H | `AcceleratorSpeedBonus +1` | 2 |
| Benzene | C₆H₆ | 6C + 6H | Unlocks **Auto-Inject** (see `docs/features/auto-injection.md`) | 3 |
| Diamond | C₁₂ | 12C | `MaxLoadBonus +15` | 3 |

**MVP total Bond Points available: 10.** The Laboratory tree (see `docs/features/laboratory.md`) costs ~31 BP — players can afford ~⅓ of it during the MVP window. Heavier-Element Bonds will fund the remainder.

Numbers are first-draft and may shift after playtesting. The catalog lives in `internal/sim/bonds.go`.

### Heavier-Element Bonds (designed, not implemented)

Bonds that require Lithium, Beryllium, Boron, Nitrogen, Oxygen, etc. are designed but gated behind those Elements gaining compound recipes. Adding them is a catalog-level change with no further architectural work needed.

## Bonds tab (UI)

Unlocked once the player owns ≥1 Token of any Element. Lives as a new tab in the right-side panel, peer to the existing inventory and codex tabs.

**MVP — list view:**

A scrollable list of synthesisable compounds. Each row shows:

- Compound name + formula ("Acetylene — C₂H₂")
- Token cost breakdown ("2× Carbon, 2× Hydrogen")
- A **Synthesise** button — enabled when the player has the Tokens, greyed otherwise.
- A small "+1 BP" / "+2 BP" / "+3 BP" badge indicating the award.
- Owned Bonds: row dimmed, button replaced with a checkmark.

A separate **Crystallise** section at the top of the Bonds tab shows each Element's reserve and the cost of its next Token, with one Crystallise button per Element. This is the only place Tokens are minted.

Hovering a Bond row shows an inline tooltip:

- One-line real-world description (e.g. "Methane — the simplest hydrocarbon").
- In-game effect summary.

Synthesis is irreversible; spent Tokens are gone.

## Bond effect integration

Bond effects plug into the Phase 6 `rebuildModifiers` pipeline (ADR 0010). The pipeline is extended to read both `PurchasedUpgrades` (paid in $USD) and `BondsState` (synthesised) and merge their effects into a single `GlobalModifiers`. See ADR 0016 for the integration shape.

Each Bond entry carries an `Apply(*GlobalModifiers)` closure, identical to a `GlobalUpgrade`. A Bond is, mechanically, a `GlobalUpgrade` with a Token cost instead of a $USD cost and a different gating predicate. The two catalogs share the modifier surface; their purchase paths are separate.

## Save compatibility

`TokenInventory map[Element]int`, `BondsState map[BondID]bool`, and `BinderReserves map[Element]int` are new fields with `omitempty`. Saves from before prestige load with empty values. No save-envelope bump.

## Related

- `internal/sim/bonds.go` — Bond catalog, `Apply` closures, `SynthesiseBond` entry point.
- `internal/sim/economy.go` — `CrystallisationCost`, `CrystalliseToken`.
- `internal/sim/state.go` — `TokenInventory`, `BondsState`, `BinderReserves`.
- `internal/ui/bonds_tab.go` — list-view UI.
- `docs/adr/0016-token-and-bond-economy.md` — storage shapes, cost function, modifier integration.
- `docs/adr/0010-global-modifier-pipeline.md` — modifier pipeline that Bonds plug into.
- `docs/features/component-binder.md` — what feeds the reserves.
- `docs/features/laboratory.md` — what Bond Points are spent on.
- `docs/features/auto-injection.md` — what the Benzene Bond unlocks.
