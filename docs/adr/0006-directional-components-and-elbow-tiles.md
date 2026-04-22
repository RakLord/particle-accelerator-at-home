# ADR 0006 — Directional components and elbow tiles

**Status:** accepted.
**Date:** 2026-04-22.

## Context

The first art pass exposed a mismatch between the simulation model and the tiles now on the board:

1. **Accelerators are drawn as in-line tube segments.** They visually imply a travel axis. A Subject entering from the side should not be able to pass through and still receive the speed bonus.
2. **The old Rotator model was direction-agnostic.** It stored only `TurnLeft` / `TurnRight` and turned any incoming Subject relative to its current heading. That worked with abstract CW / CCW icons, but not with elbow-pipe art. A physical elbow tile only connects two specific edges.
3. **Invalid entry now needs a failure mode.** The current `sim.Component` interface can only transform a `Subject`; it cannot reject and destroy one. Directional tiles need that power.
4. **The current two elbow art variants are mirrored, not mechanically distinct.** Once the tile becomes a real elbow pipe, the geometry is fully described by which two sides are connected. Mirrored art is cosmetic, not a separate gameplay concept.

## Decision

**1. Components may destroy a Subject during `Apply`.**

- `sim.Component.Apply` now returns `(Subject, lost bool)`.
- `tick.go` treats component-caused destruction exactly like any other loss event: the Subject is removed and its `Load` is freed.
- Components that never reject entry return `lost=false`.

**2. Simple Accelerators are directional by orientation.**

- `components.SimpleAccelerator` gains `Orientation sim.Direction`.
- The orientation is stored in all four cardinal directions so placement, rendering, and reconfigure all use one consistent representation.
- Gameplay is axis-based:
  - `North` and `South` accept only vertical travel.
  - `East` and `West` accept only horizontal travel.
- A Subject entering side-on is destroyed.

**3. Rotators become elbow tiles with orientation, not generic CW / CCW turners.**

- The existing `components.Rotator` type and `sim.KindRotator` catalog key stay in place to avoid unnecessary churn in inventory, costs, and save wiring.
- The gameplay model changes: `components.Rotator` gains `Orientation sim.Direction` and no longer stores a left/right turn mode.
- Orientation defines the elbow's two connected sides. The canonical unrotated elbow connects `Orientation.Left()` and `Orientation`.
  - Example: `Orientation = North` means the elbow connects the west and north edges.
- Entry rules:
  - entering from one connected side exits through the other;
  - entering from any unconnected side destroys the Subject.
- Reverse traversal is valid by construction, which matches a physical pipe elbow and keeps the rule easy to reason about.

**4. User-facing UI calls the tile `Elbow`, but inventory / cost bookkeeping stays on `KindRotator`.**

- The palette, labels, and art presentation use `Elbow`.
- Internal kind strings remain unchanged so the cost catalog and starter inventory do not need a parallel rename as part of this feature.

**5. Runtime sprite rotation is the source of truth for directional art.**

- Injector, Accelerator, and Elbow tiles all render by rotating a single authored sprite to the stored orientation.
- The mirrored elbow variant currently in `assets/images/tiles/rotator_ccw.png` is treated as non-load-bearing art. The simulation does not distinguish mirrored elbow types.

## Consequences

**Wins**
- Board visuals and gameplay rules now agree: directional tiles have directional behaviour.
- Elbow pipes become intuitive because their valid entries are exactly the openings the player can see.
- Render-side interpolation already supports quarter arcs through turning cells, so elbow motion continues to animate cleanly.

**Costs**
- `sim.Component` becomes a slightly richer interface, which touches every concrete Component and a few tests.
- Wrong-side entry is now fatal, so some previously-valid player layouts become invalid by design.
- The code keeps the legacy `Rotator` / `KindRotator` names internally for now, which is mildly confusing next to the user-facing `Elbow` label.

## Alternatives considered

- **Keep generic left/right rotators and render the elbow based on recent traffic.** Rejected: idle tiles would be visually ambiguous and live traffic would make them flicker between shapes.
- **Split elbows into separate CW and CCW gameplay kinds.** Rejected: a true elbow pipe's behaviour is already fully described by its connected edges; mirrored art does not justify a second inventory / cost / save concept.
- **Treat wrong-side entry as pass-through instead of destruction.** Rejected: it conflicts with the authored tile language and weakens directional layout design.
- **Store only horizontal / vertical on Accelerators.** Rejected: four cardinal orientations keep placement and rendering consistent with other directional components for negligible extra complexity.

## Related

- `internal/sim/components.go` — component interface now supports loss.
- `internal/sim/tick.go` — component-caused destruction handling.
- `internal/sim/components/accelerator.go` — directional axis check.
- `internal/sim/components/rotator.go` — elbow entry / exit geometry.
- `internal/input/input.go` — placement defaults and reconfigure cycling.
- `internal/render/sprites.go` — runtime sprite rotation for directional tiles.
- `internal/render/palette.go` — user-facing `Elbow` label.
