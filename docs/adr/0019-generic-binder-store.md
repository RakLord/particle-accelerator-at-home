# ADR 0019 — Generic Binder Store

**Status:** accepted (implemented for Phase 4 sim + right-panel MVP UI).
**Date:** 2026-04-25.

## Context

ADR 0015 specified a typed Binder: each placed Binder was bound to one Element and capacity scaled with the number of Binders of that Element. During implementation planning, the design changed: the physical Binder should be a simple endpoint, while capacity and reserve management should live in a separate Binder Store display.

## Decision

**1. Binder is generic.**

The placed `KindBinder` Component has no Element field. It accepts any Subject, destroys it, and asks the tick loop to bank one reserve count for `Subject.Element`.

**2. Binder Store capacity is per Element and global.**

Capacity is not per cell and does not scale with placed Binder count.

```text
EffectiveBinderStoreCapacity(element) = BaseBinderStoreCapacity[element] × BinderStoreCapacityMultiplier
```

Base capacities use the existing tuning table:

| Element | Base capacity |
|---|---:|
| Hydrogen | 15 |
| Helium | 8 |
| Lithium | 30 |
| Carbon | 100 |

**3. Capacity multiplier is a global modifier.**

`GlobalModifiers.BinderStoreCapacityMul` is the shared multiplier read by `GameState.EffectiveBinderStoreCapacity`. Dense Packing currently contributes `2^level`; future global upgrades can multiply the same field.

**4. Bonds tab appears after owning a Token.**

The Binder Store display is responsible for reserve/capacity display and Token crystallisation. The Bonds tab is responsible for synthesis and appears once `TokenInventory` contains at least one Token. It remains visible after a Bond is synthesised so owned Bonds do not disappear when the last Token is spent.

## Consequences

- Placement is simpler: no Binder Element picker and no rebinding behavior.
- Capacity is easier to explain because it appears in one Store UI rather than on grid cells.
- Physical Binders are throughput endpoints, not storage containers.
- Dense Packing remains valuable but now upgrades the global store rather than individual components.

## Related

- `docs/features/0022-component-binder.md` — updated player-facing Binder behavior.
- `docs/features/0021-bonds-and-tokens.md` — Token and Bonds UI split.
- `docs/features/0023-laboratory.md` — Dense Packing updated to Binder Store capacity.
- ADR 0015 — superseded typed-Binder model.
