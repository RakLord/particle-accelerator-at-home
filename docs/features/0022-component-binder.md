# Binder

**Status:** Phase 4.

## Behaviour

A Binder is a generic prestige endpoint — a Collector-class Component that accepts any Subject and stores it in that Element's Binder Store reserve instead of awarding $USD or research.

`(Subject) → ∅` — the Subject is removed from the grid on entry. No payout, no value awarded.

The physical Binder has no Element setting and no local storage. It only deletes the incoming Subject and adds one count to `GameState.BinderReserves[Subject.Element]` if that Element's Binder Store has capacity.

### Reserve semantics

Subjects banked by any Binder feed `GameState.BinderReserves[Element]`, a per-Element Binder Store. Capacity belongs to the store, not to the physical grid Component.

Per-Element Binder Store base capacity:

| Element | Base store capacity | Rationale |
|---|---:|---|
| Hydrogen | 15 | Gas — hard to contain. |
| Helium | 8 | Noble gas — notoriously difficult to store. |
| Lithium | 30 | Reactive metal, moderate. |
| Carbon | 100 | Solid, stable — stores well in bulk. |

Heavier-element capacities are TBD and added when those Elements gain compound recipes.

The **effective per-Element capacity** on `GameState` is:

```
EffectiveBinderStoreCapacity(e) = BaseBinderStoreCapacity[e] × BinderStoreCapacityMultiplier
```

`BinderStoreCapacityMultiplier` is a global modifier. Dense Packing currently contributes `2^DensePackingLevel`, capped at level 5 (×32), and future global upgrades can multiply the same field. See `docs/features/0023-laboratory.md`.

### At capacity — destroy

When a Subject enters a Binder and `BinderReserves[Subject.Element] >= EffectiveBinderStoreCapacity(Subject.Element)`, the Subject is **destroyed** with no $USD, no research, no refund, no Token credit. Its `Load` is freed.

Fullness is shown in the Binder Store display, not as a per-cell overlay, because capacity is not owned by the physical Binder. A future polish pass should add a one-shot "Binder Store full" notification on the first over-capacity incineration of a session (see `docs/features/0017-helper-notifications.md`).

This is harsh by design: routing Subjects into a Binder is an active commitment to bank them, and walking away from a full Binder loses Subjects. The Laboratory's Dense Packing upgrade (×2 cap per level) is the lever for shrinking that risk window.

### Unsupported Elements

A Subject whose Element has no Binder Store capacity entry is **destroyed**. No reserve gain, no payout, no warning beyond the `lost` event in the collection log. Add store capacity for heavier Elements when their Bond recipes ship.

## Placement

Binder is purchased from the inventory like any other Component (see `docs/features/0014-inventory.md`). There is no per-Binder Element picker.

In MVP, Hydrogen and Carbon reserves are the main Bond inputs. Helium and Lithium can also be stored because their base capacities are defined; heavier Elements are unsupported until their Bond recipes and store capacities ship.

## Cost

From the Component cost catalog (see `docs/features/0007-component-cost.md` for the formula):

- `Base = $40`
- `Growth = 1.30`
- Soft cap starts around the sixth purchase (Binders are deliberately scarce).

Binders sit in the same cost tier as Catalyst — they are a strategic investment, not a spammable utility.

## Design intent

Binder is the first Component whose **purpose is to sacrifice income**. Every other Component either preserves or amplifies $USD throughput. Binder voids it: a Subject routed through a Binder yields nothing on this run, in exchange for a (possibly distant) Token toward a permanent Bond.

This forces a meaningful split-grid decision on every Carbon-era run: how many Subjects do I route to income, and how many do I bank? With per-Element Token cost scaling (`docs/features/0021-bonds-and-tokens.md`), banking the *first* Token of each Element is cheap; banking the third is a real grid commitment.

The Binder also gives the prestige loop a placement-shaped surface. Without it, prestige would just be a button. With it, the player is rearranging their grid every run to feed the bank.

## Related

- `internal/sim/components/binder.go` — Component implementation.
- `internal/sim/kinds.go` — `KindBinder` registration.
- `internal/sim/state.go` — `BinderReserves`.
- `internal/sim/binder_store.go` — store capacities and `EffectiveBinderStoreCapacity`.
- `docs/adr/0019-generic-binder-store.md` — generic Binder and Store capacity model.
- `docs/features/0021-bonds-and-tokens.md` — what banked reserves are spent on.
- `docs/features/0014-inventory.md` — placement/purchase flow.
- `docs/features/0007-component-cost.md` — cost formula.
- `docs/features/0016-component-creation-and-balancing.md` — balancing checklist.
