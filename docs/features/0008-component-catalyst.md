# Catalyst

**Status:** Phase 3.

## Behaviour

A Catalyst transforms a Subject by multiplying its **Mass** by a factor, but **only if** the research level for the Subject's Element meets the Catalyst's threshold.

Below the threshold, the Catalyst is inert — the Subject passes through unchanged.

At or above the threshold, the Catalyst applies its full Mass multiplier.

### Per-Element behaviour

Catalyst reads the current Element's research level at the moment the Subject enters. A single Catalyst in a mixed build can be a no-op for Hydrogen Subjects (low research) and live for Helium Subjects (higher research) on the same tick.

### Stacking

Mass multipliers from multiple Catalysts on the same path stack **multiplicatively**. Two Catalysts in sequence at the same tier produce `bonus × bonus` on eligible Subjects.

## Design intent

Catalyst rewards **research investment retroactively**. Early-game builds place Catalysts and see no effect; the same builds become dramatically more valuable once the player pushes research past the threshold. This shortens the feedback loop between "I spent $USD unlocking research" and "my board got stronger" — the player doesn't have to tear down and rebuild to feel the upgrade.

The component pairs with heavier Elements: a Catalyst that only activates at Helium research ≥ 20 is a natural way to make late-game Elements feel distinct from early-game ones on the same board.

## Scaling with research

Below the threshold the Catalyst is inert. At or above it, the Mass multiplier scales with the Subject Element's research level:

```
mul = 1 + k · log10(research − 24)
```

At exactly `research = 25`, `log10(1) = 0` so `mul = 1.0` — the component activates but applies a unity factor. This is deliberate: the first collection past the gate is a soft on-ramp, not a sudden cliff. Each additional research point lifts the curve, and the effect compounds with whatever Mass scaling the rest of the board already applies.

## Tiers

Tierable. See `docs/features/0011-component-tiers.md`. Higher tiers steepen the curve via a larger `k`; the research activation threshold (`25`) is fixed across tiers — a tier up makes the effect stronger, not easier to activate.

| Tier | `k`  | Sample multiplier at R = 30 / 50 / 100 / 500 |
|------|-----:|---|
| T1 | 0.70 | 1.54 / 1.99 / 2.32 / 2.87 |
| T2 | 0.95 | 1.74 / 2.34 / 2.79 / 3.54 |
| T3 | 1.25 | 1.97 / 2.77 / 3.35 / 4.35 |

Tier coefficients (`catalystKByTier`) and the threshold (`catalystResearchThreshold`) both live in `internal/sim/components/catalyst.go`.

## Related

- `internal/sim/components/catalyst.go`
- `docs/adr/0008-apply-context-and-grid-view.md` — research read is what makes this component possible.
- `docs/features/0011-component-tiers.md`
- `docs/features/0004-periodic-table.md` — research progression across Elements.
- `docs/features/0001-value-formula.md` — Mass feeds collected value.
