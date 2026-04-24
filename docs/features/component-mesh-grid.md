# Mesh Grid

**Status:** Phase 2.

## Behaviour

`(Subject) → Subject` — divides the Subject's fixed-point Speed by a tier-driven divisor. Fractional output is allowed, so Mesh Grid can slow `Speed 1` to `0.5` instead of becoming inert.

| Tier | Divisor |
|---|---:|
| T1 | `÷2` |
| T2 | `÷3` |
| T3 | `÷4` |

### Fractional Speed

Speed is stored in fixed-point hundredths (`1.00` = internal `100`). Mesh Grid divides that fixed-point value and clamps any positive result that would truncate to zero up to the smallest positive Speed (`0.01`). A true `Speed 0` remains inert.

Example at T1 (divisor 2):

| Incoming Speed | Outgoing Speed |
|----------------|----------------|
| 0              | 0 (inert)      |
| 0.5            | 0.25           |
| 1              | 0.5            |
| 2              | 1              |
| 3              | 1.5            |
| 4              | 2              |
| 5              | 2.5            |
| 6              | 3              |

## Design intent

Mesh Grid is a **tool, not a trap**. The player should be able to hit thresholds in other Components' speed bands by throttling a fast Subject back down, including below `Speed 1` for heavier Elements that start slow. Tier progression turns Mesh Grid into a more aggressive throttle — at T3 a Subject at Speed 8 drops to 2, letting the player compress long accelerator chains into short ones when they need to hit low-speed-band triggers.

## Related

- `internal/sim/components/mesh_grid.go`
- `docs/features/component-tiers.md`
- `docs/features/value-formula.md` — the Speed axis feeds into collected value.
