# Component cost & inventory

**Status:** Phase 2.

## Concept

Every placeable Accelerator Component costs $USD to acquire. Cost scales with the quantity already owned of that kind, so the first purchase is cheap and subsequent ones ramp. Components stay in an **inventory** when removed from the grid — erasing a cell never refunds $USD but does return the component to the available pool.

## Formula

For each purchase of `kind`:

```
cost = ceil( Base[kind] * Growth[kind] ^ Owned[kind] * Π modifiers(state, kind) )
```

- `Base` and `Growth` are tuning numbers in `sim.ComponentCatalog`.
- `Owned[kind]` is the total number ever purchased of that kind — monotonic. Erase does not decrement it.
- The final cost is ceilinged to a whole-dollar integer — no fractional $USD ever reaches the UI or the deduction.

The current catalog is in `internal/sim/component_cost.go`.

## Inventory

- `Available[kind] = Owned[kind] - count-placed-on-grid(kind)`.
- Placement consumes one from Available; if Available is zero, placement auto-purchases (atomic: if the purchase fails for insufficient funds, placement is a no-op).
- Erasing a cell returns the component to Available automatically: Owned is unchanged and the placed count drops.
- Overwriting an occupied cell is the same as erase + place: the displaced component returns to Available (possibly a different kind), the new component consumes or auto-purchases as normal.
- Reconfigure (rotate Injector direction, flip Rotator turn) is **free**. The player already owns the component.

## Starter inventory

A brand-new game begins with enough components to build a minimal loop:

- 1 Injector
- 2 Simple Accelerators
- 1 Rotator
- 1 Collector

Starting $USD is unchanged at `0`. The player can build a Hydrogen loop, collect a few Subjects, and earn the first purchase. Starter counts live in `sim.starterInventory()` — tune there.

## Extensibility

Prestige, research, and event effects layer in through registered `CostModifier` functions:

```go
type CostModifier func(s *GameState, kind ComponentKind) bignum.Decimal
sim.RegisterCostModifier(myModifier)
```

Each modifier returns a per-state, per-kind multiplier. The final cost multiplies every registered modifier's output. Today the list is empty (no modifiers shipped), so the formula reduces to `Base * Growth^Owned` ceilinged. Phase-4 prestige upgrades land here without re-shaping the surface.

## UI

The tool palette (`internal/render/palette.go`) shows a sub-label per purchasable tool:

```
have: N · next: $X
```

Tools you can't afford *and* have zero of are dimmed (muted swatch + muted text). Clicking a dimmed tool is a no-op; nothing pays, nothing places. Element availability is handled in the Codex injection selector, not by separate Injector inventory entries.

## Save format

Additive: `GameState.Owned` is a new `map[ComponentKind]int` with `omitempty`. Save envelope stays at version 2. Saves that predate this feature load cleanly and have their Owned map seeded from the grid contents at load time, so no player loses placed components on update. See `docs/adr/0005-component-cost-and-inventory.md`.

## Related

- `internal/sim/component_cost.go` — catalog, cost, purchase, inventory queries.
- `internal/sim/kinds.go` — `ComponentKind` constants, including the synthetic `KindCollector`.
- `internal/input/input.go` — placement / erase / reconfigure wiring.
- `internal/render/palette.go` — inventory + cost sub-label per tool.
- `docs/adr/0005-component-cost-and-inventory.md` — data-model and save-format decisions.
