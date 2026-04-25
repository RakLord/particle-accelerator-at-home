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
| 0 | 5 |
| 1 | 10 |
| 2 | 15 |
| 3 | 50 |
| 4 | 50 |
| 5 | 100 |
| 6 | 200 |
| n ≥ 7 | doubles each Token |

The cost is **per-Element**: owning 2 Carbon Tokens does not affect the cost of the first Hydrogen Token. Each Element has its own scaling wall.

The cost is read from `BinderReserves[Element]` only — the player chooses to crystallise via the Binder Store UI, and the function is:

```go
func CrystalliseToken(s *GameState, e Element) error
//   precondition:  BinderReserves[e] >= CrystallisationCost(e, TokenInventory[e])
//   on success:    BinderReserves[e] -= cost; TokenInventory[e] += 1
```

Crystallisation is manual. Subjects in the Binder reserve do not auto-convert.

### Cost intent

On a first Carbon run the player should be able to mint the first few Tokens without a long idle wall, so the first three costs are intentionally light. Per-Element scaling still makes compounds that require many Tokens of the same Element (Benzene 6C+6H, Diamond 12C) multi-run goals. Heavier-Element runs are expected to bring Token-gain bonuses (planned, not in MVP) that flatten this curve further.

## Bonds — compound catalog

A Bond is a compound the player has synthesised. Synthesis spends Tokens and is permanent — once owned, a Bond never leaves `BondsState`. Each Bond:

- Applies a permanent modifier to `GameState.Modifiers` via the same closure-based pipeline as global upgrades (ADR 0010).
- Awards a one-time amount of Bond Points (see `docs/features/0023-laboratory.md`).
- Marks its row in the Bonds tab as owned.

### MVP catalog (Carbon + Hydrogen only)

| Bond | Formula | Token cost | Modifier effect | Bond Points awarded |
|---|---|---|---|---:|
| Methane | CH₄ | 1C + 4H | `CollectorValueMul ×1.15` | 1 |
| Acetylene | C₂H₂ | 2C + 2H | `InjectorRateMul ×1.333` (cooldown 25% faster) | 1 |
| Ethylene | C₂H₄ | 2C + 4H | `AcceleratorSpeedBonus +1` | 2 |
| Benzene | C₆H₆ | 6C + 6H | Unlocks **Auto-Inject** (see `docs/features/0020-auto-injection.md`) | 3 |
| Diamond | C₁₂ | 12C | `MaxLoadBonus +15` | 3 |

**MVP total Bond Points available: 10.** The Laboratory tree (see `docs/features/0023-laboratory.md`) costs ~31 BP — players can afford ~⅓ of it during the MVP window. Heavier-Element Bonds will fund the remainder.

Numbers are first-draft and may shift after playtesting. The catalog lives in `internal/sim/bonds.go`.

### Heavier-Element Bonds (designed, not implemented)

Bonds that require Lithium, Beryllium, Boron, Nitrogen, Oxygen, etc. are designed but gated behind those Elements gaining compound recipes. Adding them is a catalog-level change with no further architectural work needed.

## Bonds tab (UI)

Unlocked once the player owns ≥1 Token of any Element, and remains visible once any Bond has been synthesised. It lives in the right-side panel's Carbon Loop area. The first Token is minted from the Binder Store display; the Bonds tab is for synthesis once at least one Token exists.

**MVP — list view:**

A compact list of synthesisable compounds. Each row shows:

- Compound name + formula ("Acetylene — C₂H₂")
- Token cost breakdown ("2× Carbon, 2× Hydrogen")
- A **Synthesise** button — enabled when the player has the Tokens, greyed otherwise.
- A small "+1 BP" / "+2 BP" / "+3 BP" badge indicating the award.
- Owned Bonds: row highlighted and button labelled `Owned`.

The **Binder Store** display shows each Element's reserve, capacity, next Token cost, and one Crystallise button per Element. This is the only place Tokens are minted.

Deferred polish: hovering a Bond row should show an inline tooltip:

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
- `internal/sim/bonds.go` — `CrystallisationCost`, `CrystalliseToken`.
- `internal/sim/state.go` — `TokenInventory`, `BondsState`, `BinderReserves`.
- `internal/render/prestige_panel.go` — Binder Store and Bonds list UI.
- `docs/adr/0016-token-and-bond-economy.md` — storage shapes, cost function, modifier integration.
- `docs/adr/0010-global-modifier-pipeline.md` — modifier pipeline that Bonds plug into.
- `docs/features/0022-component-binder.md` — what feeds the reserves.
- `docs/features/0023-laboratory.md` — what Bond Points are spent on.
- `docs/features/0020-auto-injection.md` — what the Benzene Bond unlocks.
