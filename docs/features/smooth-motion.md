# Smooth motion: interpolation, rotator arcs, and trails

**Status:** live. Replaces the "tick-granular rendering" note in `docs/overview.md` — the Phase-3 render-side interpolation TODO is now closed.

## What changed

Previously, Subjects snapped to their logical cell each tick. `DefaultTickRate=10 Hz` was a workaround for the resulting teleport visual. This feature adds:

1. A sim-side **SpeedDivisor** so displayed `Speed=1` traverses one cell every 10 ticks (≈ 1 s) instead of one per tick — the previous speed felt too fast even when smoothed.
2. A render-side **alpha** (wall-clock fraction within the current sim tick) used to interpolate Subjects between ticks, including fractional in-cell glide during the 9 of 10 ticks where no cell boundary is crossed.
3. **Quarter-arc rendering through rotator cells** so subjects curve around turns instead of L-cutting through the cell center.
4. A toggleable **particle trail** (default on) that leaves a fading dot behind each Subject.

## Sim contract

`sim.SpeedDivisor = 10`. Speed is fixed-point hundredths: internal `100` displays as `Speed=1`. Every tick, each Subject accumulates its fixed-point `Speed` into `StepProgress`; every `sim.StepProgressPerCell` of accumulated progress corresponds to one cell of movement. A Subject with displayed `Speed=1` crosses one cell every 10 ticks; `Speed=10` crosses one per tick; `Speed=20` crosses two.

Speed's meaning in the collection-value formula (`internal/sim/economy.go`) is **unchanged** at the UI level — `collectValue` converts fixed-point `s.Speed` back to the displayed value before multiplying. SpeedDivisor is purely a movement-rate divisor; economy math is orthogonal.

`stepSubject` (`internal/sim/tick.go`) snapshots the per-tick motion state before advancing:

- `PrevPosition` ← current `Position`
- `PrevInDirection` ← `InDirection` (how the Subject arrived at its current cell)
- `PrevStepProgress` ← `StepProgress`
- `Path` is reset to `[Position]`, then appended with each new cell entered this tick

`InDirection` is updated at every boundary crossing, capturing the pre-Apply movement direction so the renderer knows how the Subject arrived even after a rotator changes `Direction`.

**Invariant:** no `Component.Apply` implementation may touch `Path`, `StepProgress`, `PrevStepProgress`, `InDirection`, or `PrevInDirection`. Apply is declared `(Subject) -> Subject` by-value; the returned copy shares the `Path` slice header, and overwriting these motion fields would break render interpolation.

All motion-snapshot fields are tagged `json:"-"` — the save format is unchanged. On load, `InDirection` is defaulted to `Direction` (see `sim.Load`) so the first post-load frame doesn't render a spurious arc.

## Render contract

Wall-clock alpha (`render.Game.tickAlpha`):

```
alpha = clamp01( time.Since(lastTickAt) / tickDuration )
```

`lastTickAt` is set immediately after each `state.Tick()` in `Update`. `tickDuration = time.Second / state.TickRate`.

`subjectPixel(sub, alpha)` (`internal/render/subject_motion.go`) walks `sub.Path` consuming virtual progress (`alpha × sub.Speed`), with the first cell starting at the `PrevStepProgress` fraction. It emits pixel coordinates from `cellInternalPos`, which resolves a cell + inbound/outbound direction + fraction to:

- **Straight cell** (`dirIn == dirOut`): line from inbound edge midpoint to outbound edge midpoint (passing through cell center at `t=0.5`).
- **Turn cell** (`dirIn != dirOut`): quarter arc from inbound edge midpoint to outbound edge midpoint, centered on the inner corner of the turn, radius `cellSize/2`.

The edge-midpoint parametrization (rather than center-to-center) is deliberate: it makes boundary crossings continuous (outbound midpoint of cell A == inbound midpoint of cell B) and makes turn arcs tangent to the straight segments on either side of a rotator.

Subjects spawn with `StepProgress = SpeedDivisor/2` so they appear visually centered on the injector cell on their first render frame. Cost: the first cell takes half as long as subsequent cells; the offset persists but isn't visually jarring.

## Rotator arc geometry

For a turn at cell `C` with inbound direction `D_in` and outbound direction `D_out`:

- Arc center = the cell corner shared by the `opposite(D_in)` edge and the `D_out` edge (the "inner" corner of the turn).
- Radius = `cellSize/2`.
- Sweep = 90° from the inbound edge midpoint (angle `atan2(ix-innerX, iy-innerY)`) to the outbound edge midpoint, shortest direction.

Example (East → South, right turn): inbound-side = West edge; outbound-side = South edge; inner corner = SW; arc sweeps from (cellLeft, cellMidY) to (cellMidX, cellBottom).

## Particle trails

Stored as `[]trailSample` on `render.Game`, purely session-scoped. Each `Draw`:

- When `ui.TrailsEnabled` is true, every live Subject's interpolated position is appended as a new sample.
- All existing samples age by one Draw frame; expired samples (`Age >= trailLifetime = 45` frames, ≈ 0.75 s at 60 FPS) are compacted out.
- Samples draw as small circles with alpha = `1 - Age/trailLifetime`, beneath the live Subjects.

No identity tracking — samples are positional snapshots, not per-Subject trails. Collisions, deaths, and collections don't need special handling.

Toggles (both do the same thing, including clearing the buffer so old samples don't linger for their lifetime after disable):

- Hotkey **T** (global, handled in `Game.handleInput`).
- Settings modal checkbox row "Particle trails (T)".

The flag lives on `ui.UIState` (default `true`) and is **not persisted**. If stickiness is wanted later, move it onto `GameState` under a `Prefs` struct.

## Related files

- `internal/sim/subject.go` — motion-state field definitions
- `internal/sim/tick.go` — `SpeedDivisor`, `stepSubject`
- `internal/sim/save.go` — on-load default for `InDirection`
- `internal/sim/components/injector.go` — spawn-centering `StepProgress`
- `internal/render/subject_motion.go` — `subjectPixel` + geometry helpers
- `internal/render/trail.go` — trail buffer + draw
- `internal/render/game.go` — `tickAlpha`, `T` hotkey wiring
- `internal/render/settings.go` — checkbox row

## Follow-ups (intentionally deferred)

- Raise `sim.DefaultTickRate` from 10 to 60. The `state.go` comment pointing at this is now unblocked by interpolation; it's deferred because 60 Hz with Speed=1 crosses a 5×5 grid in ~80 ms even interpolated — a gameplay change, not a rendering one.
- Persist `TrailsEnabled` across sessions if users ask for stickiness.
- Replace the colored-circle trail samples with a connected ribbon once we have a shader pipeline.
