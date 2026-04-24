# Catalyst

**Status:** Phase 3.

## Behaviour

A Catalyst transforms a Subject by multiplying its **Mass** by a factor, but **only if** the research level for the Subject's Element meets the Catalyst's threshold.

Below the threshold, the Catalyst is inert — the Subject passes through unchanged.

At or above the threshold, the Catalyst applies its full Mass multiplier.

### Per-Element behaviour

Catalyst reads the current Element's research level at the moment the Subject enters. A single Catalyst in a mixed build can be a no-op for Hydrogen Subjects (low research) and live for Helium Subjects (higher research) on the same tick.

### Stacking

Mass multipliers from multiple Catalysts on the same path stack **multiplicatively**. Two Catalysts in sequence at the same tier produce `bonus × bonus` on eligible Subjects.

## Design intent

Catalyst rewards **research investment retroactively**. Early-game builds place Catalysts and see no effect; the same builds become dramatically more valuable once the player pushes research past the threshold. This shortens the feedback loop between "I spent $USD unlocking research" and "my board got stronger" — the player doesn't have to tear down and rebuild to feel the upgrade.

The component pairs with heavier Elements: a Catalyst that only activates at Helium research ≥ 20 is a natural way to make late-game Elements feel distinct from early-game ones on the same board.

## Tiers

Tierable. See `docs/features/component-tiers.md`. Higher tiers increase the Mass multiplier. The research threshold itself is fixed per Catalyst (a tier up makes the effect stronger, not easier to activate).

## Related

- `internal/sim/components/catalyst.go`
- `docs/adr/0008-apply-context-and-grid-view.md` — research read is what makes this component possible.
- `docs/features/component-tiers.md`
- `docs/features/periodic-table.md` — research progression across Elements.
- `docs/features/value-formula.md` — Mass feeds collected value.
