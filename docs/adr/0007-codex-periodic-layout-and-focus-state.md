# ADR 0007 — Codex periodic layout and focused stat card

**Status:** accepted.
**Date:** 2026-04-23.

## Context

The first Codex implementation shipped as a small row-based modal. That was enough to prove the Helium unlock flow, but it leaves two problems once the Codex becomes a real progression surface:

1. A row list does not visually read as a periodic table, so it undersells the Element fantasy.
2. Hover-driven inspection and click-to-pin interaction need transient focus state, but that state should not leak into simulation or persistence.

The redesign also adds per-Element best stats. Those are game progress and should be saved, but they are distinct from the Codex's moment-to-moment interaction state.

## Decision

**1. The Codex is a large overlay with real periodic-table positions.**

- Elements render in an 18-column frame using canonical `Period` and `Group` coordinates.
- Early versions of the table will be sparse on purpose. Empty space is part of the presentation, not a bug.
- The row-based list is removed rather than maintained alongside the table.

**2. Element spatial metadata lives in `sim.ElementInfo`.**

- `AtomicNumber`, `Period`, and `Group` are part of the game catalog, not a render-only side table.
- This keeps unlock/progression data and element identity in one place.
- Render consumes that metadata to place tiles; future features can reuse it without duplicating tables.

**3. Focus interaction state lives in `render.Game`, not `ui.UIState` or `sim.GameState`.**

- Hovered element and pinned element are transient session UI state.
- They are not persisted and are not relevant outside rendering/input routing.
- `ui.UIState` keeps coarse modal flags (`CodexOpen`, `SettingsOpen`) but does not grow a dependency on `sim.Element`.

**4. The unlock CTA lives on the focused stat card, not on the table tile.**

- Tiles stay visually simple: number, symbol, and state tint.
- The centered stat card owns detailed text, best stats, and the unlock button.
- This preserves the periodic-table silhouette while keeping progression actions explicit.

## Consequences

**Wins**

- The Codex now reads as a real periodic table even with a tiny catalog.
- Hover inspection and click/tap fallback both fit naturally without persisting ephemeral UI state.
- Best stats have a clear home on the card instead of bloating tiles or lists.

**Costs**

- The screen uses more dedicated layout code than the old row list.
- Future Elements must provide valid periodic metadata when added to `ElementCatalog`.
- Hit-testing and focus resolution become more stateful than the old single-button-row model.

## Alternatives considered

- **Compact "periodic-table inspired" layout.** Rejected: it solves early sparse space but loses the immediate recognition of true positions.
- **Keep the small modal and just reskin rows.** Rejected: the underlying interaction remains a list, not a table.
- **Store hovered/pinned element on `ui.UIState`.** Rejected: `ui` is intentionally lightweight and currently free of `sim` types.
- **Put the unlock button on the tile itself.** Rejected: tiles become crowded and the table stops reading as a table.

## Related

- `docs/features/periodic-table.md` — gameplay-facing Codex behavior and stat semantics.
- `internal/sim/economy.go` — element catalog metadata.
- `internal/render/periodic_table.go` — layout, hit-testing, and stat-card rendering.
- `internal/render/game.go` — Codex input flow and transient focus state.
