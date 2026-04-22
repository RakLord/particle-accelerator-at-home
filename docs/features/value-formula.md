# Value formula (on collection)

**Status:** MVP placeholder — revisit in Phase 2/3.

## Current (Phase 1 MVP)

```
$USD awarded = Mass × Speed × elementMultiplier(Element)
```

Magnetism is accepted by the function signature but multiplied by 0, so that Phase 2's Magnetiser does not require a signature change across saves and tests.

Element multipliers (Phase 1 has Hydrogen only):

| Element   | Multiplier |
|-----------|------------|
| Hydrogen  | 1.0        |

## Deferred

- Per-Element **research level** multiplier (Phase 2).
- **Global upgrade** multipliers (Phase 3).
- Non-zero **Magnetism** coefficient (Phase 2, with Magnetiser).
- Balancing pass once multiple Elements exist.

## Design constraint (from `docs/overview.md`)

> Each input axis should feel individually meaningful — no single axis dominating.

The MVP formula deliberately collapses to `Mass × Speed` because Phase 1 only has those two axes to vary. Proper balancing waits for Phase 2 (speed bands + Magnetism + research) where the multi-axis trade-off actually exists.
