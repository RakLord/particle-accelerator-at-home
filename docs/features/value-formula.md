# Value formula (on collection)

**Status:** live — Phase 2 constants in place. Balance pass still pending.

## Current formula

```
value = (Mass × Speed × speedK + Magnetism × magK) × elementMultiplier(Element) × collectorValueMul
```

Constants (see `internal/sim/economy.go`):

| Name     | Value | Notes                          |
|----------|-------|--------------------------------|
| `speedK` | 1.0   | Gain on the Mass × Speed axis. |
| `magK`   | 0.5   | Gain on the Magnetism axis.    |

`collectorValueMul` comes from `GlobalModifiers.CollectorValueMul` (Normalized to 1 when unset).

Element multipliers:

| Element  | Symbol | Base multiplier |
|----------|--------|-----------------|
| Hydrogen | H      | 1.0             |
| Helium   | He     | 2.5             |

## Deferred

- **Global upgrade** multipliers (Phase 3 — cross-cutting $USD sinks).
- Per-Component tiers beyond the current `+1` (Phase 3).
- **Research-driven value modulation** is planned as an upgrader-applied effect that transforms a Subject mid-flight, not as a collection-time multiplier. Research still accrues on collection and still gates element/tier unlocks; it no longer feeds directly into payout.
- Balancing pass once Phase 2 content is live. Expect `magK` to move.

## Design constraint (from `docs/overview.md`)

> Each input axis should feel individually meaningful — no single axis dominating.

The formula keeps Mass, Speed, and Magnetism independently multiplicative or additive so that a build optimising any one of them still pays off. The per-Element multiplier is the progression lever.
