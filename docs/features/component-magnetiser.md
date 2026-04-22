# Magnetiser

**Status:** Phase 2.

## Behaviour

`(Subject) → Subject` — adds `Bonus` to the Subject's `Magnetism`. MVP default `Bonus = 1`.

### Speed band

Magnetiser triggers when `Speed >= 1`. This is effectively always-on for moving Subjects, but the gate is structured so a future "supercharged" Magnetiser could require a minimum speed threshold without reshuffling the component interface.

## Stacking

Magnetism stacks additively across multiple Magnetisers on a path. Collection bakes Magnetism into the value formula with coefficient `magK = 0.5` (see `docs/features/value-formula.md`).

## Design intent

Magnetism is the second independent axis introduced in Phase 2. Pairing a Magnetiser path with a long Simple Accelerator chain lets the player choose whether to optimise for the Speed axis (raw cells/tick) or the Magnetism axis (collected value multiplier). The coefficient is deliberately lower than the Speed axis gain so Magnetism alone can't outrun a Speed build.

## Related

- `internal/sim/component_magnetiser.go`
- `docs/features/value-formula.md`
