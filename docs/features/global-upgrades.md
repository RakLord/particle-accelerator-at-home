# Global upgrades

**Status:** Phase 3.

## Concept

Global upgrades are one-shot $USD purchases that apply a cross-cutting bonus to the whole board. Unlike component purchases (which add to inventory) or tier upgrades (which strengthen one component kind), a global upgrade is a named unlock that reshapes the economy once bought and is never repurchased.

Examples:
- "All Collectors yield +10% $USD."
- "Inject cooldown recovers twice as fast."
- "All Accelerators grant +1 flat Speed."
- "All component purchases cost ×0.8."
- "Max Load +8."

## Purchase

Each global upgrade has:
- A **name** and a short **description** of its effect.
- A **$USD cost**.
- A **research prerequisite** — typically a specific Element at a specific research level.
- A purchased / unpurchased **state**. There is no quantity; once bought, that's it.

The player sees global upgrades in a dedicated shop tab listed with all requirements. Greyed-out entries can't be bought yet (research unmet or insufficient funds); purchased entries are marked as owned and disabled.

## Effect

Every purchased upgrade contributes to an aggregated `GlobalModifiers` state. Hot paths (collection value, manual injection cooldown, component Apply behaviour, `MaxLoad` enforcement) read the aggregated state on every tick. Stacked upgrades compose:

- **Multiplicative** upgrades (e.g. Collector value, component purchase cost, Injector rate, Magnetiser bonus) multiply over each other. Two `+10%` Collector upgrades yield `1.10 × 1.10 = 1.21×` total; two `×0.8` component-cost discounts would yield `0.64×` cost.
- **Additive** upgrades (e.g. flat Speed bonus, flat Max Load bonus) sum over each other.

The split is per-field, documented in ADR 0010.

## First drop

| Upgrade | Effect | Prerequisite |
|---|---|---|
| Collector Coils I | Collector value `×1.10` | Hydrogen research ≥ 5 |
| Rapid Injection | Injection cooldown rate `×2` | Hydrogen research ≥ 10 |
| Power Surge | Accelerators grant `+1` flat Speed | Hydrogen research ≥ 15 |
| Capacity Expansion | Max Load `+8` | Helium unlocked |

Tuning numbers are subject to playtest; the catalog lives in `internal/sim/global_upgrades.go`.

## Interaction with tiers

Global upgrades stack with component tiers. A Tier-3 Accelerator (+3 Speed baseline from the tier table) with the Power Surge upgrade grants `+4` Speed. Neither subsystem replaces the other.

## Save compatibility

Purchased upgrades persist as a set of upgrade IDs on `GameState`. The aggregated modifier state is derived from that set on every save load and on every purchase — retuning an upgrade's numeric effect in a future release applies to all existing saves on next load. See ADR 0010 for derivation and save-schema details.

Saves from before global upgrades exist load into a state with no upgrades purchased and identity modifiers — same gameplay as pre-upgrade.

## Related

- `internal/sim/global_upgrades.go` — catalog and purchase entry point.
- `internal/sim/modifiers.go` — aggregated modifier state.
- `internal/ui/upgrades_tab.go` — shop UI.
- `docs/adr/0010-global-modifier-pipeline.md` — data model, derivation rule, save-compat.
- `docs/features/component-cost.md` — where component purchase cost modifiers apply.
- `docs/features/component-creation-and-balancing.md` — practical guide for component cost tuning.
- `docs/features/component-tiers.md` — orthogonal progression axis.
- `docs/features/value-formula.md` — where Collector value modifiers apply.
