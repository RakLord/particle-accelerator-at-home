# Mesh Grid

**Status:** Phase 2.

## Behaviour

`(Subject) → Subject` — integer-divides the Subject's Speed by a tier-driven divisor. The per-tier divisor and band floor live in `meshGridDivisorByTier` / `meshGridMinSpeedByTier` in `internal/sim/components/mesh_grid.go`.

| Tier | Divisor | Min speed band |
|---|---|---|
| T1 | `÷2` | Speed ≥ 2 |
| T2 | `÷3` | Speed ≥ 3 |
| T3 | `÷4` | Speed ≥ 4 |

### Speed band

Mesh Grid only triggers when `Speed` is at or above the tier's band floor. Below the floor the component is inert. This prevents the degenerate case where a low-Speed Subject gets floored to 0 and becomes trapped on the cell, unable to leave. The band floor rises with tier so higher tiers can't trap medium-speed Subjects either.

Example at T1 (divisor 2, band ≥ 2):

| Incoming Speed | Outgoing Speed |
|----------------|----------------|
| 0              | 0 (inert)      |
| 1              | 1 (inert)      |
| 2              | 1              |
| 3              | 1              |
| 4              | 2              |
| 5              | 2              |
| 6              | 3              |

## Design intent

Mesh Grid is a **tool, not a trap**. The player should be able to hit thresholds in other Components' speed bands by throttling a fast Subject back down. Tier progression turns Mesh Grid into a more aggressive throttle — at T3 a Subject at Speed 8 drops to 2, letting the player compress long accelerator chains into short ones when they need to hit low-speed-band triggers.

## Related

- `internal/sim/components/mesh_grid.go`
- `docs/features/component-tiers.md`
- `docs/features/value-formula.md` — the Speed axis feeds into collected value.
