# Component cost & inventory

**Status:** Phase 2.

## Concept

Every placeable Accelerator Component costs $USD to acquire. Cost scales with the quantity already owned of that kind, so the first purchase is cheap and subsequent ones ramp. Components stay in an **inventory** when removed from the grid — erasing a cell never refunds $USD but does return the component to the available pool.

## Formula

For each purchase of `kind`, first compute the raw exponential cost:

```
raw = Base[kind] * Growth[kind] ^ Owned[kind]
```

Then apply the optional soft cap:

```
if SoftCapAt[kind] is set and raw > SoftCapAt[kind]:
    shaped = SoftCapAt[kind] * (raw / SoftCapAt[kind]) ^ SoftCapPower[kind]
else:
    shaped = raw
```

Then apply multipliers and round up:

```
cost = ceil( shaped * GlobalComponentCostMultiplier * Π modifiers(state, kind) )
```

- `Base`, `Growth`, optional `SoftCapAt`, and optional `SoftCapPower` are tuning numbers in `sim.ComponentCatalog`.
- `Owned[kind]` is the total number ever purchased of that kind — monotonic. Erase does not decrement it.
- `SoftCapAt` is a price threshold, not an owned-count threshold. It lets early purchases follow the normal curve, then makes pushing far past the intended quantity aggressively more expensive.
- `SoftCapPower` should usually be an integer `2` or `3`. `0`/`1` means no extra soft-cap amplification.
- `GlobalComponentCostMultiplier` is `GameState.Modifiers.ComponentCostMul`, normalized to `1` today. Future global upgrades can set it below `1` for discounts or above `1` for challenge modes.
- Registered `CostModifier` functions still multiply after the soft cap and global multiplier for future prestige/research/event effects.
- The final cost is ceilinged to a whole-dollar integer — no fractional $USD ever reaches the UI or the deduction.

The current catalog is in `internal/sim/component_cost.go`.

### Current catalog tuning

| Kind | Base | Growth | SoftCapAt | SoftCapPower | Early notes |
|---|---:|---:|---:|---:|---|
| Injector | `$10` | `1.15` | unset | unset | Utility source. |
| Accelerator | `$5` | `3.2` | `$5K` | `2` | Early costs: `$5`, `$16`, `$52`, `$164`, `$525`, `$1 678`; soft-cap pressure starts around `$5K`. |
| Mesh Grid | `$15` | `1.20` | unset | unset | Speed-band utility. |
| Magnetiser | `$100` | `5` | unset | unset | Second purchase is `$500`. |
| Rotator | `$8` | `1.15` | unset | unset | Routing utility. |
| Collector | `$50` | `1.25` | unset | unset | Endpoint/economy sink. |
| Resonator | `$50` | `1.35` | unset | unset | Slightly steeper than early utility curves. |
| Catalyst | `$1 000` | `12` | unset | unset | Second purchase is `$12K`. |
| Duplicator | `$10 000` | `125` | unset | unset | Second purchase is `$1.25M`. |

## Inventory

- `Available[kind] = Owned[kind] - count-placed-on-grid(kind)`.
- Placement consumes one from Available; if Available is zero, placement auto-purchases (atomic: if the purchase fails for insufficient funds, placement is a no-op).
- Erasing a cell returns the component to Available automatically: Owned is unchanged and the placed count drops.
- Overwriting an occupied cell is the same as erase + place: the displaced component returns to Available (possibly a different kind), the new component consumes or auto-purchases as normal.
- Reconfigure (rotate Injector direction, flip Rotator turn) is **free**. The player already owns the component.

## Starter inventory

A brand-new game begins with only the two endpoints of an acceleration loop:

- 1 Injector
- 1 Collector

Everything else — accelerators, elbows, mesh grids, magnetisers, etc. — must be purchased. Starting $USD is unchanged at `0`, so the player's first move is to place the Injector and Collector adjacently; the first collected Subject funds the first Accelerator ($5). Starter counts live in `sim.starterInventory()` — tune there.

## Extensibility

Prestige, research, and event effects layer in through registered `CostModifier` functions:

```go
type CostModifier func(s *GameState, kind ComponentKind) bignum.Decimal
sim.RegisterCostModifier(myModifier)
```

Each modifier returns a per-state, per-kind multiplier. The final cost multiplies every registered modifier's output after the built-in soft-cap shaping and global `ComponentCostMul`. Today the list is empty (no registered modifiers shipped), so the formula reduces to the catalog curve plus any `GameState.Modifiers.ComponentCostMul`. Phase-4 prestige upgrades land here without re-shaping the surface.

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
- `docs/features/0016-component-creation-and-balancing.md` — workflow for adding new components and choosing cost curves.
