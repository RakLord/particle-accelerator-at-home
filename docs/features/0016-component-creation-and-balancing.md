# Component Creation And Cost Balancing

**Status:** living guide.

Use this guide when adding a new Accelerator Component or retuning an existing one. It connects the simulation implementation, UI copy, inventory economics, and cost-curve balancing rules in one place.

## Required References

- `docs/overview.md` — canonical game model and terminology.
- `docs/features/0007-component-cost.md` — exact component purchase formula.
- `docs/features/0011-component-tiers.md` — tier model and tier-upgrade pricing.
- Relevant existing component feature docs in `docs/features/component-*.md`.
- ADRs for component architecture: `docs/adr/0008-apply-context-and-grid-view.md`, `docs/adr/0009-subject-emitter-capability.md`, and `docs/adr/0011-component-tier-primitive.md`.

## Adding A Component

1. Define or reuse a `sim.ComponentKind` in `internal/sim/kinds.go`.
2. Implement the component as a pure Subject transform in `internal/sim/components/`.
3. Register save construction through the component registry path used by `internal/sim/save.go`.
4. Add tests in `internal/sim/components/` for the component's tick/apply behavior and save round-trip shape.
5. Add inventory/tool mapping in `internal/ui/lock.go`.
6. Add player-facing copy in `internal/ui/component_info.go`.
7. Add or wire sprites in `internal/render/sprites.go` and render fallbacks where needed.
8. Add purchase-cost tuning in `internal/sim/component_cost.go`.
9. If tierable, add tier stat tables near the component implementation and tier upgrade entries in `internal/sim/tier.go`.
10. Update or create a feature doc in `docs/features/`.

Keep `internal/sim` headless. Do not import Ebitengine, render, or UI packages from simulation code.

## Purchase Cost Formula

Component purchase costs use this shape:

```
raw = Base * Growth ^ Owned

if SoftCapAt is set and raw > SoftCapAt:
    shaped = SoftCapAt * (raw / SoftCapAt) ^ SoftCapPower
else:
    shaped = raw

cost = ceil(shaped * ComponentCostMul * registeredCostModifiers)
```

Where:

- `Owned` is total owned for that component kind, not currently placed count.
- `Base` is the first purchase cost when `Owned == 0`.
- `Growth` controls the normal exponential ramp.
- `SoftCapAt` is a price threshold where the curve starts punishing over-buying.
- `SoftCapPower` controls soft-cap severity. Use `2` for a strong warning, `3` for an aggressive wall.
- `ComponentCostMul` is the normalized global multiplier on `GameState.Modifiers`; it defaults to `1`.
- Registered cost modifiers are for future prestige/research/event effects.

## Choosing Base And Growth

Start from the first three purchases. Pick the first purchase from intended unlock timing, then solve growth by the second or third desired purchase.

Example for Accelerator:

```
Base = 5
Target second purchase = 15-20
Target third purchase = 40-70
Growth = 3.2
```

Result before soft cap:

| Owned Before Purchase | Cost |
|---:|---:|
| 0 | `$5` |
| 1 | `$16` |
| 2 | `$52` |
| 3 | `$164` |
| 4 | `$525` |
| 5 | `$1 678` |

If purchase two is too cheap, raise `Growth`. If purchase one is wrong, change `Base`; do not distort `Growth` to fix the first purchase.

## Choosing Soft Caps

Use soft caps when the player should feel pushed toward a different progression axis instead of buying the same component forever.

Good soft-cap candidates:

- Components that directly multiply income or path power.
- Components that should be strong in small numbers but not spammed.
- Late-game or phase-4 components with high layout impact.

Poor soft-cap candidates:

- Basic routing utility required to make the grid function.
- Components whose quantity is naturally constrained by grid area.
- Very early tutorial components unless the cap is high enough to stay invisible during onboarding.

Recommended starting values:

| Component Role | SoftCapAt | SoftCapPower |
|---|---:|---:|
| Early power component | `$5K-$25K` | `2` |
| Economy multiplier | `$25K-$100K` | `2` |
| Late-game power spike | `$100K+` | `2-3` |
| Utility/routing | unset | unset |

Accelerator currently uses `SoftCapAt=$5K`, `SoftCapPower=2`.

Current high-impact catalog examples:

| Component | Base | Growth | Second Purchase |
|---|---:|---:|---:|
| Magnetiser | `$100` | `5` | `$500` |
| Catalyst | `$1 000` | `12` | `$12K` |
| Duplicator | `$10 000` | `125` | `$1.25M` |
| Resonator | `$50` | `1.35` | `$68` |

## Graphing Curves

When balancing, graph at least these columns for owned counts `0-20`:

- `raw = Base * Growth^Owned`.
- `shaped`, after soft cap.
- `final`, after global/component modifiers.
- cumulative spend, if the goal is pacing over several purchases.

Do not judge a curve from the first purchase alone. The important shape is usually purchases 2-8 for early components and purchases 5-15 for late components.

## Tier Upgrade Costs Are Separate

Component purchase costs and tier upgrade costs are independent axes.

- Buying another Accelerator uses `internal/sim/component_cost.go`.
- Upgrading all Accelerators to T2/T3 uses `internal/sim/tier.go`.

Do not compensate for an expensive purchase curve by silently changing tier upgrade costs in the same change unless the feature request explicitly covers both.

## Tests To Update

When changing cost curves:

- Update `internal/sim/component_cost_test.go` expected values.
- Add a targeted test if using a new soft-cap shape or global modifier behavior.
- Keep whole-dollar ceiling behavior covered.

When adding a new component:

- Add component behavior tests.
- Add save tests if the component has custom fields.
- Add UI catalog coverage if new tool mappings are introduced.
