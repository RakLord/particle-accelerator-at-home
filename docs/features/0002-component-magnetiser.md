# Magnetiser

**Status:** Phase 2.

## Behaviour

`(Subject) → Subject` — adds a tier-driven flat bonus to the Subject's `Magnetism`. The per-tier bonus lives in the `magnetiserBonusByTier` table in `internal/sim/components/magnetiser.go`.

| Tier | Magnetism bonus |
|---|---|
| T1 | `+1` |
| T2 | `+2` |
| T3 | `+3` |

### Speed band

Magnetiser triggers when `Speed >= 1`. This is effectively always-on for moving Subjects, but the gate is structured so a future "supercharged" Magnetiser could require a minimum speed threshold without reshuffling the component interface.

## Stacking

Magnetism stacks additively across multiple Magnetisers on a path. Collection bakes Magnetism into the value formula with coefficient `magK = 0.5` (see `docs/features/0001-value-formula.md`).

The `GlobalModifiers.MagnetiserBonusMul` multiplier (set by global upgrades — see `docs/features/0012-global-upgrades.md`) multiplies the per-pass bonus before it's added to the Subject.

## Design intent

Magnetism is the second independent axis introduced in Phase 2. Pairing a Magnetiser path with a long Simple Accelerator chain lets the player choose whether to optimise for the Speed axis (raw cells/tick) or the Magnetism axis (collected value multiplier). The coefficient is deliberately lower than the Speed axis gain so Magnetism alone can't outrun a Speed build.

## Related

- `internal/sim/components/magnetiser.go`
- `docs/features/0011-component-tiers.md`
- `docs/features/0012-global-upgrades.md`
- `docs/features/0001-value-formula.md`
