# Compressor

**Status:** Phase 3.

## Behaviour

A Compressor transforms a Subject by multiplying its **Mass** by the inverse of its displayed **Speed**, scaled by a per-tier coefficient.

At or above Speed `1.00`, the Compressor is inert — the Subject passes through unchanged.

Below `1.00`, the multiplier is:

```
mul = TierCoef × (1 / displayed_speed)
```

Examples at T1 (`TierCoef = 1`):

| Speed | Mass multiplier |
|---:|---:|
| `1.00` | `×1` (no-op) |
| `0.50` | `×2` |
| `0.25` | `×4` |
| `0.10` | `×10` |

Zero-speed Subjects are also treated as a no-op — they cannot be rewarded via division-by-zero.

### Stacking

Multiple Compressors on the same path stack **multiplicatively**. Two T1 Compressors in a row on a Subject at Speed `0.10` yield a `×100` Mass multiplier combined.

## Design intent

Compressor is the natural pay-off for **deliberately throttled paths**. Mesh Grid exists to divide Speed; without a component that rewards slow Subjects, it's only ever a tool for re-entering activation bands. Compressor inverts that: the more a player commits to the slow lane, the more they earn back.

This creates a tension against the speed-seeking components (Accelerator, Resonator) and gives mixed builds a real layout decision rather than "always go faster".

## Tiers

Tierable. Higher tiers increase the coefficient applied on top of the `1/Speed` ratio.

| Tier | Coefficient | Multiplier at Speed 0.10 |
|---:|---:|---:|
| T1 | `×1.0` | `×10` |
| T2 | `×1.5` | `×15` |
| T3 | `×2.0` | `×20` |

Coefficient table lives in `internal/sim/components/compressor.go` (`compressorCoefByTier`).

## Cost curve

Purchase cost grows aggressively — the third copy sits on the soft-cap and the fourth punishes over-buying hard:

| Owned before purchase | Price |
|---:|---:|
| 0 | `$7,000` |
| 1 | `$105,000` |
| 2 | `$1,575,000` (at soft cap) |
| 3 | `$354,375,000` |

Tuning lives in `ComponentCatalog[KindCompressor]` in `internal/sim/component_cost.go`.

## Related

- `docs/overview.md` — canonical game model.
- `docs/features/component-mesh-grid.md` — primary synergy partner.
- `docs/features/component-catalyst.md` — the other Mass multiplier.
- `docs/features/component-tiers.md` — tier model and purchase flow.
