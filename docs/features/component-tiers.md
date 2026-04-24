# Component tiers

**Status:** Phase 3.

## Concept

Every tierable component kind has a **global tier level** shared across every placed instance of that kind. Upgrading a tier once upgrades every Accelerator you've already placed, every Accelerator currently in inventory, and every Accelerator you'll ever place — the upgrade is per-kind, not per-instance.

Tier 1 is the baseline. Higher tiers produce stronger effects (more Speed, more Magnetism, steeper Mesh Grid reduction, etc.).

## Which components tier

| Kind | Tierable? | Stat progression |
|---|---|---|
| Simple Accelerator | yes | Speed bonus grows per tier |
| Magnetiser | yes | Magnetism bonus grows per tier |
| Mesh Grid | yes | Speed divisor grows per tier (tier 2 divides by 3, tier 3 by 4) |
| Resonator | yes | Per-neighbour Speed contribution grows per tier |
| Catalyst | yes | Mass multiplier grows per tier |
| Duplicator | yes | Per-output Mass fraction grows per tier (T1 is mass-conservative; T2+ actively creates mass) |
| Injector | no | Per-instance config (element, direction, interval) instead |
| Rotator / Elbow | no | Per-instance orientation instead |
| Collector | no | Governed by `docs/features/global-upgrades.md` |

## How tiers are purchased

Each tierable kind has an ordered list of tier unlocks in the shop: T2, then T3, then T4, and so on. Each unlock is gated by:

- `$USD` cost (rising with tier).
- A per-Element research threshold — typically Hydrogen research for early tiers, heavier Elements for later tiers.
- The previous tier must already be unlocked. T3 cannot be bought before T2.

A purchase advances the global tier by exactly one level for that kind. The advance is immediate and applies to every placed and inventoried instance on the next tick.

## Starter state

A fresh game starts every kind at Tier 1. Starter inventory amounts (see `docs/features/component-cost.md`) are unchanged by tiering — the player always receives Tier 1 items at game start, regardless of how many tier upgrades they later purchase.

## Interaction with buying more components

Purchase cost (`Base * Growth^Owned` — see `docs/features/component-cost.md`) does **not** change with tier. Buying your 20th Accelerator costs the same whether the Accelerator tier is T1 or T3. Tier upgrades and bulk-purchase costs are independent axes.

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
- `docs/features/global-upgrades.md` — orthogonal flat-bonus axis.
