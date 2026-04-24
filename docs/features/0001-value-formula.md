# Value formula (on collection)

**Status:** live — Phase 2 constants in place. Balance pass still pending.

## Current formula

```
value = (Mass × Speed × speedK + Magnetism × magK) × elementMultiplier(Element) × collectorValueMul
```

Injected `Mass` and `Speed` come from the selected Element's catalog entry. `BaseMass` uses standard atomic mass values, so heavier Elements pay more through the same Mass axis as Component effects. `BaseSpeed` is a coarse gameplay approximation of equal-energy acceleration: Hydrogen and Helium start faster, while heavier Elements start slower until Components accelerate them.

Simulation Speed is fixed-point hundredths. UI and economy code use the displayed value, so internal `100` is `Speed 1.00`, internal `50` is `Speed 0.50`, and the value formula reads those as `1.0` and `0.5` respectively.

Constants (see `internal/sim/economy.go`):

| Name     | Value | Notes                          |
|----------|-------|--------------------------------|
| `speedK` | 1.0   | Gain on the Mass × Speed axis. |
| `magK`   | 0.5   | Gain on the Magnetism axis.    |

`collectorValueMul` comes from `GlobalModifiers.CollectorValueMul` (Normalized to 1 when unset).

Element multipliers:

| Element | Symbol | Base Mass | Base Speed | Base multiplier |
|---|---|---:|---:|---:|
| Hydrogen | H | 1.008 | 2 | 1.0 |
| Helium | He | 4.003 | 2 | 1.5 |
| Lithium | Li | 6.94 | 1 | 1.8 |
| Beryllium | Be | 9.012 | 1 | 2.1 |
| Boron | B | 10.81 | 1 | 2.4 |
| Carbon | C | 12.011 | 1 | 2.8 |
| Nitrogen | N | 14.007 | 1 | 3.2 |
| Oxygen | O | 15.999 | 1 | 3.6 |
| Fluorine | F | 18.998 | 1 | 3.8 |
| Neon | Ne | 20.180 | 1 | 4.0 |
| Sodium | Na | 22.990 | 1 | 4.5 |
| Magnesium | Mg | 24.305 | 1 | 5.0 |
| Aluminium | Al | 26.982 | 1 | 5.7 |
| Silicon | Si | 28.085 | 1 | 6.5 |
| Phosphorus | P | 30.974 | 1 | 7.2 |
| Sulfur | S | 32.06 | 1 | 8.0 |
| Chlorine | Cl | 35.45 | 1 | 8.6 |
| Argon | Ar | 39.948 | 1 | 9.0 |
| Potassium | K | 39.098 | 1 | 10.0 |
| Calcium | Ca | 40.078 | 1 | 12.0 |

## Deferred

- **Global upgrade** multipliers (Phase 3 — cross-cutting $USD sinks).
- Per-Component tiers beyond the current `+1` (Phase 3).
- **Research-driven value modulation** is planned as an upgrader-applied effect that transforms a Subject mid-flight, not as a collection-time multiplier. Research still accrues on collection and still gates element/tier unlocks; it no longer feeds directly into payout.
- Balancing pass once Phase 2 content is live. Expect `magK` to move.

## Design constraint (from `docs/overview.md`)

> Each input axis should feel individually meaningful — no single axis dominating.

The formula keeps Mass, Speed, and Magnetism independently multiplicative or additive so that a build optimising any one of them still pays off. The per-Element multiplier is the progression lever.
