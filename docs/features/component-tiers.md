# Component tiers

**Status:** Phase 3.

## Concept

Every tierable component kind has a **global tier level** shared across every placed instance of that kind. Upgrading a tier once upgrades every Accelerator you've already placed, every Accelerator currently in inventory, and every Accelerator you'll ever place — the upgrade is per-kind, not per-instance.

Tier 1 is the baseline. Higher tiers produce stronger effects (more Speed, more Magnetism, steeper Mesh Grid reduction, etc.).

## Which components tier

| Kind | Tierable? | T1 | T2 | T3 |
|---|---|---|---|---|
| Simple Accelerator | yes | `+1` Speed | `+2` | `+3` |
| Magnetiser | yes | `+1` Magnetism | `+2` | `+3` |
| Mesh Grid | yes | `÷2` Speed | `÷3` | `÷4` |
| Resonator | yes | `+1` Speed per neighbour | `+2` | `+3` |
| Catalyst | yes | Mass `× (1 + k·log₁₀(R−24))` when R ≥ 25; `k = 0.70` | `k = 0.95` | `k = 1.25` |
| Duplicator | yes | Mass `×0.5` per output | `×0.6` | `×0.75` |
| Injector | no | Per-instance config (direction, interval); Element is selected globally in the Codex |
| Rotator / Elbow | no | Per-instance orientation |
| Collector | no | Governed by `docs/features/global-upgrades.md` |

## How tiers are purchased

Each tierable kind has an ordered list of tier unlocks in the shop: T2, then T3, then T4, and so on. Each unlock is gated by:

- `$USD` cost (rising with tier).
- A per-Element research threshold — typically Hydrogen research for early tiers, heavier Elements for later tiers.
- The previous tier must already be unlocked. T3 cannot be bought before T2.

A purchase advances the global tier by exactly one level for that kind. The advance is immediate and applies to every placed and inventoried instance on the next tick.

Initial catalog (`internal/sim/tier.go`):

| Kind | T2 cost / research | T3 cost / research |
|---|---|---|
| Simple Accelerator | $500 / Hydrogen ≥ 3 | $5 000 / Hydrogen ≥ 15 |
| Magnetiser | $800 / Hydrogen ≥ 5 | $8 000 / Hydrogen ≥ 20 |
| Mesh Grid | $400 / Hydrogen ≥ 4 | $4 000 / Hydrogen ≥ 18 |
| Resonator | $1 200 / Hydrogen ≥ 8 | $12 000 / Hydrogen ≥ 25 |
| Catalyst | $2 000 / Hydrogen ≥ 30 | $20 000 / Helium ≥ 10 |
| Duplicator | $3 000 / Hydrogen ≥ 12 | $30 000 / Helium ≥ 15 |

Values subject to playtest.

## Starter state

A fresh game starts every kind at Tier 1. Starter inventory amounts (see `docs/features/component-cost.md`) are unchanged by tiering — the player always receives Tier 1 items at game start, regardless of how many tier upgrades they later purchase.

## Interaction with buying more components

Purchase cost (catalog exponential curve plus optional soft-cap shaping — see `docs/features/component-cost.md`) does **not** change with tier. Buying your 20th Accelerator costs the same whether the Accelerator tier is T1 or T3. Tier upgrades and bulk-purchase costs are independent axes.

## Interaction with global upgrades

Global upgrades (`docs/features/global-upgrades.md`) layer on top of tier bonuses. A Tier-2 Accelerator with a "+1 Speed" global upgrade gives `+3` Speed (tier 2 gives `+2`, global adds `+1`). Global upgrades are additive flat bonuses; tiers set the baseline.

## UI

Tier upgrades live in a shop panel separate from the component-purchase inventory. Each tierable kind shows:

- Current tier.
- Next tier stat preview.
- Cost and research prerequisite.
- Greyed out when prerequisite unmet; disabled when already at max tier.

## Save compatibility

Saves from before tiers exist load as Tier 1 for every kind. No migration. See `docs/adr/0011-component-tier-primitive.md`.

## Related

- `internal/sim/tier.go` — tier model and purchase flow.
- `docs/adr/0011-component-tier-primitive.md` — data model and save-compat decisions.
- `docs/features/component-cost.md` — orthogonal bulk-purchase cost axis.
- `docs/features/component-creation-and-balancing.md` — practical guide for tuning purchase curves versus tier costs.
- `docs/features/global-upgrades.md` — orthogonal flat-bonus axis.
