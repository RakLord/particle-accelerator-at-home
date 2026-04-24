# ADR 0005 — Component cost, inventory, and additive save fields

**Status:** accepted.
**Date:** 2026-04-22.

## Context

Phase 2 needs an economic sink for $USD beyond Element unlocks. Until now, placing an Accelerator Component has been free — every tool click just mutates the grid. That makes $USD only useful for Element unlocks and defers core incremental-game pressure to later phases.

Three coupled decisions fall out of adding a component purchase cost:

1. **Where does "how many do you own" live?** Owned count is a per-kind integer on game state. The simplest model is also the most flexible.
2. **What key does the Collector use?** Collectors are stored today as `cell.IsCollector = true`, not as a `Component` instance, so they fall outside the component registry.
3. **How do old saves load?** ADR 0004 set a precedent of "envelope version bump ⇒ reject". This feature's shape change is additive, and that precedent was for a breaking serialization shape, not additive fields.

## Decision

**1. Owned is a monotonic counter on `GameState`.**

- `GameState.Owned map[ComponentKind]int` tracks the total number ever purchased per kind.
- Purchase increments; placement and erase do not touch it.
- Available (placeable right now) is always derived: `Owned[kind] - count-placed-on-grid(kind)`.
- One counter instead of two (placed + available) removes a whole class of drift bugs and keeps the save smaller.

**2. Cost is a compositional curve, not a fixed table.**

- Initial implementation used `cost = ceil(Base * Growth^Owned * Π registered CostModifiers)`.
- Current implementation keeps the same compositional surface but allows each catalog entry to add soft-cap shaping: `raw = Base * Growth^Owned`; above `SoftCapAt`, `shaped = SoftCapAt * (raw / SoftCapAt)^SoftCapPower`; final cost is `ceil(shaped * ComponentCostMul * Π registered CostModifiers)`.
- Modifiers are a package-level slice (`RegisterCostModifier(fn)`) so prestige/research/event upgrades compose without re-shaping the signature.
- Each modifier takes both the `GameState` and the `ComponentKind`, so per-kind effects ("Injectors 20% cheaper after research tier 3") fit the surface directly.

**3. Collector uses a synthetic `KindCollector = "collector"` catalog key.**

- The key is **not** registered with `componentRegistry`. Registrations would let something call `newComponentByKind("collector")` and get back nothing coherent — collectors aren't `Component` instances.
- `kinds.go` documents this invariant at the declaration.
- A cleaner long-term fix is promoting Collector to a real `Component` implementation (drop the `cell.IsCollector` bool, add an identity-Apply Component). That reshapes the save format and `tick.go`'s collection check, so it's a real v3 — deferred.

**4. `Kind*` constants live in `internal/sim`, not `internal/sim/components`.**

- `component_cost.go` needs to reference every kind's catalog entry from `internal/sim`. With the constants in the `components` subpackage (which imports `sim`), the catalog would create an import cycle.
- Constants moved up; the `components` package continues to own registration and `Apply` behaviour.

**5. Save envelope stays at version 2.**

- Adding `Owned map[ComponentKind]int` with `omitempty` is **additive**: old saves deserialize without the field, new saves include it, nothing breaks structurally.
- On load, a nil `Owned` triggers a one-time migration that seeds the map from whatever is on the grid. Long-time players don't lose the components they've already placed.
- This refines ADR 0004's "bump ⇒ reject" precedent: that rule applies to **shape-breaking** changes (field-type swaps, encoding changes). Additive fields with a safe default can land under the existing version.

## Consequences

**Wins**
- $USD becomes load-bearing for Phase 2 gameplay, not just an unlock currency.
- The modifier surface absorbs Phase-4 prestige effects (and any intermediate research/event upgrades) without reshaping.
- Old saves keep loading; existing players don't lose placed components.

**Costs**
- `Kind*` constants now live in two conceptual homes (the sim package holds the strings; the components subpackage holds the types and behaviour). Migration is mechanical but touches every component file.
- `KindCollector` is a registry-exempt special case. A reader who assumes every `ComponentKind` round-trips through the registry will hit surprises; the declaration has a comment calling this out, but it remains a trap until Collector becomes a real Component.
- Additive saves mean a future shape-breaking change still needs its own version bump. ADR 0004 is not relaxed — only refined.

## Alternatives considered

- **Track placed and available as two separate maps.** Rejected: double-bookkeeping risks drift when placement / erase / overwrite all touch state. A single counter plus a scan of the (small) grid is simpler.
- **One modifier hook returning a single scalar.** Rejected: foreseeable upgrades are per-kind, not global, so a `kind`-less signature pushes conditional branches into a god function.
- **Bump the save envelope to v3.** Rejected: this change doesn't break the envelope's shape. Rejecting all v2 saves would discard live player progress for a benign additive field.
- **Make Collector a real Component now.** Rejected: the refactor touches `cell.IsCollector`, `tick.go`'s collection check, the save format, and UI. Worth doing, but not as part of the cost feature; tracked as a follow-up.

## Related

- `internal/sim/kinds.go` — all `ComponentKind` constants, including `KindCollector`.
- `internal/sim/component_cost.go` — catalog, `ComponentCost`, `PurchaseComponent`, `CanPurchase`, modifier registry.
- `internal/sim/save.go` — migration for saves that predate the `Owned` field.
- `internal/input/input.go` — placement auto-purchase flow and overwrite semantics.
- `docs/features/component-cost.md` — gameplay-facing feature description.
- `docs/features/component-creation-and-balancing.md` — current tuning workflow for component curves and soft caps.
- ADR 0002 — versioned save schema.
- ADR 0004 — canonical save strings; this ADR refines its "bump ⇒ reject" rule to apply only to shape-breaking changes.
