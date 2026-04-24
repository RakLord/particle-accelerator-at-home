# Duplicator

**Status:** Phase 3.

## Behaviour

A Duplicator is a **T-junction**: a Subject enters one side and leaves as **two** Subjects heading out the perpendicular sides. The original Subject is consumed by the junction; the two emitted Subjects carry copies of the original's `Element`, `Speed`, `Magnetism`, and `Load` (possibly adjusted — see Stat split below), with `Direction` set to each of the junction's two output sides.

### Shape

Like the Rotator/Elbow (ADR 0006), a Duplicator has an `Orientation`. For Duplicator, `Orientation` names the **input** side; the two output sides are perpendicular (`Orientation.Left()` and `Orientation.Right()`).

- `Orientation = West` means a Subject must enter from the west edge; outputs head North and South.
- `Orientation = East` means input from the east; outputs North and South.
- `Orientation = North` or `South` means input from that side; outputs East and West.
- A Subject entering from an output side (or from the closed opposite side) is destroyed — the junction is not bidirectional.

### Stat split

Each emitted Subject receives a fraction of the incoming Subject's `Mass`, determined by the Duplicator's global tier. Speed and Magnetism are copied unchanged on both outputs — duplication costs mass, not momentum.

| Tier | Per-output Mass fraction | Total Mass leaving |
|---|---|---|
| T1 | `× 0.5` | `1.0×` (conservation) |
| T2 | `× 0.6` | `1.2×` |
| T3 | `× 0.75` | `1.5×` |

T1 is mass-conservative: a Duplicator is a pure parallelism tool at baseline, paying for itself only through two independent modifier paths or Load-pressure effects. Higher tiers actively *create* mass, turning Duplicator into a direct $USD multiplier. This is deliberately how the progression feels: T1 is a positioning tool, T2+ is an economic engine.

Numbers above are illustrative and may shift at implementation time. Tier table lives in `internal/sim/components/duplicator.go`.

### Load and grid admission

Each emitted Subject costs its own `Load` against `MaxLoad`. If the grid is full, one output may fire and the other may be silently dropped — same policy as manual Injector admission. A full grid can therefore squeeze a Duplicator's effective output to one or zero subjects.

## Design intent

Duplicator is the first component that **grows subject count**. The existing economy is bounded by how fast Injectors fire; Duplicator uncouples throughput from spawn rate, turning the grid into an amplifier rather than a pipeline. Combined with a long chain of Magnetisers on each output side, a single Injector can feed a substantial late-game build.

At T1 the mass-conservation rule keeps Duplicator honest: total collected value from the two outputs equals what one Subject would have collected if sent directly, assuming identical Collector paths — value comes from *parallelism* or from *Load pressure*. Tier upgrades break conservation in the player's favour, so tiering a Duplicator is a direct output-multiplier unlock — a high-value $USD sink. Future global upgrades (`docs/features/0012-global-upgrades.md`) can layer an additional mass-retention bonus on top of the tier fraction.

## Related

- `internal/sim/components/duplicator.go`
- `docs/adr/0009-subject-emitter-capability.md` — `Splitter` interface makes this component possible.
- `docs/adr/0006-directional-components-and-elbow-tiles.md` — shared orientation model.
- `docs/features/0011-component-tiers.md`
- `docs/features/0001-value-formula.md`
