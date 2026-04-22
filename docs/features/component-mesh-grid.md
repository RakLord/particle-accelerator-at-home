# Mesh Grid

**Status:** Phase 2.

## Behaviour

`(Subject) → Subject` — integer-halves the Subject's Speed: `Speed /= 2`.

### Speed band

Mesh Grid only triggers when `Speed >= 2`. At Speed = 1 the component is inert. This prevents the degenerate case where a Speed = 1 Subject gets floored to 0 and becomes trapped on the cell, unable to leave.

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

Mesh Grid is a **tool, not a trap**. The player should be able to hit thresholds in other Components' speed bands by throttling a fast Subject back down. The Phase 2 roster has no components with upper speed limits yet, but the mechanism is in place for Phase 3.

## Related

- `internal/sim/component_mesh_grid.go`
- `docs/features/value-formula.md` — the Speed axis feeds into collected value.
