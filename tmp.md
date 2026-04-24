# Carbon Prestige Layer — Refined Plan (Decisions Locked)

## Overview

When Carbon is unlocked, a second progression axis opens alongside the existing Genesis loop. The player can place **Binder** components that bank Subjects into an element-specific reserve. Banked Subjects can be crystallised into **Tokens**. Tokens are spent in the new **Bonds** tab to synthesise molecular compounds, each granting a permanent bonus. Completing at least one compound unlocks the prestige (reset) button.

---

## 1. Binder Component

A new Accelerator Component placed on the grid. When a Subject enters it:
- The Subject is absorbed into the Binder's internal reserve (no $USD, no research awarded).
- If the reserve is **at capacity**, the Subject is **destroyed** — gone, no refund.

This makes the trade-off sharp: routing Subjects into a Binder sacrifices all income and research from those Subjects. The player must decide how many Subjects to divert and when.

**Per-element capacity (starting values):**
| Element | Capacity | Rationale |
|---|---:|---|
| Hydrogen | 15 | Gas — hard to contain |
| Helium | 8 | Noble gas — notoriously difficult to store |
| Lithium | 30 | Reactive metal, moderate |
| Carbon | 100 | Solid, stable — stores well in bulk |
| (heavier elements) | TBD — implement when elements added | |

Capacity is a meaningful time gate: wanting Token 3 for Carbon (cost: 300) with a 100-cap Binder means filling and draining it at least three times. Capacity upgrades in the Laboratory tree make this less painful on later runs.

---

## 2. Tokens — Per-Element Crystallisation

Every element has its own token type. The crystallisation cost counter is **per element** — owning 2 Carbon Tokens does not affect the cost of your first Hydrogen Token.

**Cost sequence (per element):**
| Tokens already owned (this element) | Cost to crystallise next |
|---:|---:|
| 0 | 10 subjects |
| 1 | 50 subjects |
| 2 | 300 subjects |
| 3 | 1 500 subjects |
| 4 | 8 000 subjects |
| … | ~5× each step |

**Intent:** on a first Carbon run the player should realistically afford 1–2 tokens *per element* at most. Per-element scaling means compounds that require multiple tokens of the same element (e.g. Benzene: 6C) are genuine multi-run goals, not first-run achievements.

**Global difficulty modifier:** `GlobalModifiers.TokenCostMul` (float, default `1.0`). Stored on `GameState.Modifiers` now; a future next-layer upgrade tree can reduce it. Costs are multiplied by this value before display and before crystallisation.

---

## 3. Bonds Tab (UI)

Unlocked once the player owns at least 1 token of any element. Lives as a new tab in the right-side panel.

**Layout — MVP (list view):**

A scrollable list of synthesisable compounds. Each row:
- Compound name + formula ("Acetylene — C₂H₂")
- Token cost breakdown ("2× Carbon Token, 2× Hydrogen Token")
- **Synthesise** button — enabled when tokens are sufficient, grayed otherwise
- Owned compounds: row dimmed, button replaced with a checkmark

Hovering (or tapping) any row shows an inline tooltip:
- One-line real-world description
- In-game bonus granted

Synthesis is permanent and irreversible. Tokens are consumed on synthesis.

**Initial compound catalog (Carbon + Hydrogen only — MVP scope):**
| Compound | Formula | Token cost | In-game bonus |
|---|---|---|---|
| Methane | CH₄ | 1C + 4H | Global $USD yield ×1.15 |
| Acetylene | C₂H₂ | 2C + 2H | Injection cooldown ×0.75 |
| Ethylene | C₂H₄ | 2C + 4H | All Accelerator Speed +1 (flat) |
| Benzene | C₆H₆ | 6C + 6H | Passive income tick — earns $USD once per second without injection |
| Diamond | C×12 | 12C | Max Load +15, persists across prestige |

*Numbers are first-draft — tuning after playtesting. Benzene (6C + 6H) is a deliberate multi-run target. Diamond (12C) is an endgame flex.*

Heavier-element compounds (Lithium, Beryllium, etc.) are designed but not implemented until those elements are added to the game.

---

## 4. Prestige (Reset)

**Unlock condition:** at least one compound synthesised.

**On prestige:**
- **Resets:** $USD, grid layout, component inventory, element research, all Binder reserves, all Tokens.
- **Persists:** synthesised compounds (Bonds) and their bonuses. These are permanent.

A **Laboratory** tab appears after the first prestige, showing all owned compounds and their active bonuses.

---

## 5. Laboratory Upgrade Tree (post-prestige)

Upgrades cost **Bond Points** — 1 awarded per compound synthesised, ever. Spend them here.

| Upgrade | Max purchases | Effect | Cost per purchase |
|---|---:|---|---:|
| Covalent Memory | 1 | Start next run with Helium already unlocked | 1 |
| Stable Isotope | 1 | Research carries over at 30% on reset | 2 |
| Chain Reaction | 1 | Injectors start at 2× base emission rate | 2 |
| Carbon Core | 1 | Binder + Bonds tab unlock immediately on run start | 1 |
| Dense Packing | 5 | All Binder capacities ×2 per purchase (stacks multiplicatively, cap at 5) | 1 / 2 / 3 / 4 / 5 |

*Dense Packing at max (5 purchases): ×32 capacity. Carbon Binder cap goes from 100 → 3 200. This makes Diamond (12C = 12 × 300-cost tokens at owned=2… well, it scales, but eventually feasible).*

---

## 6. Resolved Design Decisions

| Question | Decision |
|---|---|
| Token cost counter — shared or per-element? | **Per-element.** Each element has its own scaling wall. |
| Binder full behaviour | **Destroy the Subject.** No refund. |
| Binder income / research | **None.** The entire Subject value is sacrificed. |
| Hydrogen/Helium capacity | Keep as designed; Laboratory "Dense Packing" upgrade (×2 per purchase, cap 5) covers it. |
| Element scope for MVP | **Carbon + Hydrogen only.** Heavier elements deferred until added to Genesis. |

---

## 7. Implementation Sketch

**New sim concepts (all in `internal/sim/`):**
- `TokenInventory map[Element]int` on `GameState` — tokens owned per element
- `CrystallisationCost(element Element, owned int) BigInt` — per-element cost function using the 5× sequence
- `BondsState map[CompoundID]bool` on `GameState` — synthesised compounds (survives reset)
- `BondPoints int` on `GameState` — meta-currency (survives reset)
- `GlobalModifiers.TokenCostMul float64` — difficulty multiplier, default 1.0
- Binder capacity table: `BinderCapacity map[Element]int`, modified by `Dense Packing` level from Laboratory tree

**New component:**
- `KindBinder` in `internal/sim/kinds.go`
- `internal/sim/components/binder.go` — absorbs Subject into reserve; destroys on full
- No $USD or research awarded on absorption

**New UI:**
- `internal/ui/bonds_tab.go` — compound list, synthesise logic, token display
- `internal/ui/laboratory_tab.go` — post-prestige compound log + upgrade tree

**Prestige reset function:**
- `ResetGenesis(state *GameState)` in `internal/sim/` — zeroes $USD, grid, inventory, research, Binder reserves, TokenInventory; leaves BondsState, BondPoints, Laboratory upgrade levels intact

**Save schema:**
- `TokenInventory`, `BondsState`, `BondPoints`, `LaboratoryUpgrades` are new top-level fields on the save struct
- Existing saves load with zero values (no tokens, no bonds) — no migration needed
