# Binder

**Status:** Phase 4.

## Behaviour

A Binder is a **typed endpoint** — a Collector-class Component that accepts Subjects of a specific Element and stores them in a per-Element reserve instead of awarding $USD or research.

`(Subject) → ∅` — the Subject is removed from the grid on entry. No payout, no value awarded.

Each Binder is bound to a single Element (chosen at placement time, like Injector binds to the Codex selector). A Subject of any other Element entering a Binder is **destroyed** the same way it would be at a wrong-axis Pipe — wrong tool for the job. This keeps players from accidentally banking Hydrogen into a Carbon Binder.

### Reserve semantics

Subjects banked by a Binder feed `GameState.BinderReserves[Element]`, a per-Element pool. Multiple Binders of the same Element pool into the same reserve — placing two Hydrogen Binders gives `2 × HydrogenCapacity` total banking room.

Per-Element capacity (per-Binder) at base:

| Element | Per-Binder capacity | Rationale |
|---|---:|---|
| Hydrogen | 15 | Gas — hard to contain. |
| Helium | 8 | Noble gas — notoriously difficult to store. |
| Lithium | 30 | Reactive metal, moderate. |
| Carbon | 100 | Solid, stable — stores well in bulk. |

Heavier-element capacities are TBD and added when those Elements gain compound recipes.

The **effective per-Element capacity** on `GameState` is:

```
EffectiveBinderCapacity(e) = count(Binders placed for e) × BaseBinderCapacity[e] × DensePackingMultiplier
```

`DensePackingMultiplier` is `2^DensePackingLevel` from the Laboratory tree, capped at level 5 (×32). See `docs/features/0023-laboratory.md`.

### At capacity — destroy

When a Subject enters a Binder of its own Element and `BinderReserves[Element] >= EffectiveBinderCapacity(Element)`, the Subject is **destroyed** with no $USD, no research, no refund, no Token credit. Its `Load` is freed.

A "Binder full" notification fires on the first incineration of a session (see `docs/features/0017-helper-notifications.md`). The Binder cell renders a "FULL" overlay while at capacity. Subsequent incinerations are silent — the player has been warned.

This is harsh by design: routing Subjects into a Binder is an active commitment to bank them, and walking away from a full Binder loses Subjects. The Laboratory's Dense Packing upgrade (×2 cap per level) is the lever for shrinking that risk window.

### Wrong-Element entry

A Subject whose `Element` does not match the Binder's bound Element is **destroyed**. Same outcome as wrong-axis Pipe entry. No reserve gain, no payout, no warning beyond the `lost` event in the collection log.

## Placement

Binder is purchased from the inventory like any other Component (see `docs/features/0014-inventory.md`). At placement, the player selects which Element the Binder binds to from a per-Binder picker — the same picker pattern as the Injector's Codex element selector.

A Binder cannot be re-bound after placement. To change its Element, the player removes and replaces the Component.

In MVP, only Carbon and Hydrogen Binders are useful — those are the only Elements with Bond recipes. Heavier-Element Binders are placeable but bank into a reserve with no Token recipe yet. Element pickers grey out unsupported Elements until heavier Bonds ship.

## Cost

From the Component cost catalog (see `docs/features/0007-component-cost.md` for the formula):

- `Base = $40`
- `Growth = 1.30`
- Soft cap at 5 owned (Binders are deliberately scarce).

Binders sit in the same cost tier as Catalyst — they are a strategic investment, not a spammable utility.

## Design intent

Binder is the first Component whose **purpose is to sacrifice income**. Every other Component either preserves or amplifies $USD throughput. Binder voids it: a Subject routed through a Binder yields nothing on this run, in exchange for a (possibly distant) Token toward a permanent Bond.

This forces a meaningful split-grid decision on every Carbon-era run: how many Subjects do I route to income, and how many do I bank? With per-Element Token cost scaling (`docs/features/0021-bonds-and-tokens.md`), banking the *first* Token of each Element is cheap; banking the third is a real grid commitment.

The Binder also gives the prestige loop a placement-shaped surface. Without it, prestige would just be a button. With it, the player is rearranging their grid every run to feed the bank.

## Related

- `internal/sim/components/binder.go` — Component implementation.
- `internal/sim/kinds.go` — `KindBinder` registration.
- `internal/sim/state.go` — `BinderReserves`, `EffectiveBinderCapacity`.
- `docs/adr/0015-binder-component.md` — storage model, full-behavior dispatch, capacity registry.
- `docs/features/0021-bonds-and-tokens.md` — what banked reserves are spent on.
- `docs/features/0014-inventory.md` — placement/purchase flow.
- `docs/features/0007-component-cost.md` — cost formula.
- `docs/features/0016-component-creation-and-balancing.md` — balancing checklist.
