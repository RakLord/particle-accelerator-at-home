# Pipe

**Status:** Phase 2.

## Behaviour

`(Subject) → Subject` — a straight pass-through. The Pipe stores an `Orientation` which defines its open axis (vertical or horizontal). A Subject entering along either end of that axis exits out the other end, preserving direction, Speed, Mass, and Magnetism. A Subject entering perpendicular to the open axis is destroyed.

Implementation: `internal/sim/components/pipe.go`.

| Entry side | Horizontal pipe (`Orientation=East`) | Vertical pipe (`Orientation=North`) |
|---|---|---|
| East  | pass-through (continues west → east) | destroyed |
| West  | pass-through (continues east → west) | destroyed |
| North | destroyed | pass-through (continues south → north) |
| South | destroyed | pass-through (continues north → south) |

## Reconfiguration

Scrolling (or left-click rotate) cycles the `Orientation` through all four cardinal directions, matching the convention of the other orientable components. North/South produce the same vertical sprite and East/West the same horizontal sprite, so users effectively see two visual states.

## Cost

From `internal/sim/component_cost.go`:

- `Base = $8`
- `Growth = 1.15`
- No soft cap — routing utility is deliberately unconstrained (see `docs/features/0016-component-creation-and-balancing.md` §Choosing Soft Caps).

Pipe uses the same $8 base as Rotator (Elbow) but a gentler growth curve (`1.15` vs Rotator's `1.20`), so Pipe stays cheap to spam for long runs while Rotators remain the slightly-premium turn tile.

## Design intent

Before Pipe, the only way to carry a Subject across several cells was to use Accelerators, Mesh Grids, or Magnetisers as improvised routing. Those components all mutate the Subject, so layouts that only needed to *move* a particle had to over-invest in stat changes or spam Elbows. Pipe fills that gap as a cheap, inert straight segment — it's the "wire" of the grid.

## Related

- `internal/sim/components/pipe.go`
- `internal/sim/component_cost.go` — cost catalog entry
- `docs/features/0007-component-cost.md` — cost formula
- `docs/features/0016-component-creation-and-balancing.md` — balancing workflow
