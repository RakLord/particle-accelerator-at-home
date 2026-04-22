# Value formula (on collection)

**Status:** live — Phase 2 constants in place. Balance pass still pending.

## Current formula

```
value = (Mass × Speed × speedK + Magnetism × magK) × elementMultiplier(Element) × (1 + research / researchK)
```

Constants (see `internal/sim/economy.go`):

| Name        | Value | Notes                                       |
|-------------|-------|---------------------------------------------|
| `speedK`    | 1.0   | Gain on the Mass × Speed axis.              |
| `magK`      | 0.5   | Gain on the Magnetism axis.                 |
| `researchK` | 50    | Research count that doubles the multiplier. |

`research` is snapshotted **before** the collection's own increment, so the first subject of a new Element earns the base multiplier.

Element multipliers:

| Element  | Symbol | Base multiplier |
|----------|--------|-----------------|
| Hydrogen | H      | 1.0             |
| Helium   | He     | 2.5             |

## Deferred

- **Global upgrade** multipliers (Phase 3 — cross-cutting $USD sinks).
- Per-Component tiers beyond the current `+1` (Phase 3).
- Balancing pass once Phase 2 content is live. Expect `magK` and `researchK` to move.

## Design constraint (from `docs/overview.md`)

> Each input axis should feel individually meaningful — no single axis dominating.

The Phase 2 formula keeps the three gameplay axes (Speed, Magnetism, research) independently multiplicative or additive so that a build optimising any one of them still pays off. The per-Element multiplier is the fourth knob and is the lever progression uses.
