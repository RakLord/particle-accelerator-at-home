# Periodic Table (Codex)

**Status:** Phase 2.

The Codex is a large overlay reached from the "Codex" button in the header. It presents Elements in a visually recognizable periodic-table layout, then shows detail in a centered stat card for the currently focused Element.

## Layout

The table uses real periodic-table coordinates, not a compact list.

- Each Element in `sim.ElementCatalog` supplies display metadata: `AtomicNumber`, `Period`, and `Group`.
- The Codex renders an 18-column table frame so early-game Elements can sit in their canonical positions even when the rest of the table is still empty.
- Current shipped positions:

| Element  | AtomicNumber | Period | Group |
|----------|--------------|--------|-------|
| Hydrogen | 1            | 1      | 1     |
| Helium   | 2            | 1      | 18    |

Sparse early-game layouts are intentional. H in the top-left and He in the top-right should read as a real periodic table, not as missing UI.

## Tile contents

Each occupied element tile shows:

- atomic number
- symbol
- a state tint reflecting `Unlocked`, `Purchasable`, or `Research-locked`

Tiles do **not** contain the unlock CTA. The action belongs to the stat card.

## Interaction

The Codex supports both hover and click/tap interaction.

1. **Hover** focuses the tile under the cursor.
2. **Click/tap** pins a tile so its card remains visible without hover.
3. Clicking the pinned tile again unpins it.
4. If a tile is pinned, the pinned Element is the active focus; otherwise the hovered Element is the active focus.

If no Element is focused, the overlay shows the table and a short hint to hover or select an Element.

## Stat card

The focused Element opens a centered stat card above the table. The card is the primary detail surface and shows:

- atomic number
- symbol and name
- unlock state
- research count
- element multiplier (static per-Element base multiplier used by `collectValue`)
- injection action for unlocked Elements
- best stats:
  - max Speed
  - max Mass
  - max collected value

The symbol/name block should be visually dominant so the card reads more like an inspected element than a spreadsheet row.

## Best-stat semantics

Best stats are tracked per Element and persisted on `GameState`.

- `max Speed` updates only when a Subject of that Element is collected.
- `max Mass` updates only when a Subject of that Element is collected.
- `max collected value` updates from the actual payout awarded on collection.

These are deliberately **collection-time** stats, not peak in-flight telemetry. The Codex records the best results the player has successfully converted into progress.

## Unlock flow

Each Element's unlock state is driven by two predicates:

1. **Research gate** — `Research[UnlocksFrom] >= ResearchThreshold`. While below, the card shows that the Element is locked and how much prerequisite research remains.
2. **Purchase gate** — once the research gate is clear, the card shows an `Unlock for $N` button. Clicking deducts `UnlockCost` from `USD` and flips `GameState.UnlockedElements[e] = true`.

Unlocking an Element makes it selectable as the global injection Element. All Injector components emit whichever unlocked Element is selected here; changing the selection affects existing Injectors immediately.

Current catalog (`internal/sim/economy.go`):

| Element  | UnlocksFrom | ResearchThreshold | UnlockCost |
|----------|-------------|-------------------|------------|
| Hydrogen | —           | 0 (seeded unlocked) | 0        |
| Helium   | Hydrogen    | 10                | 500        |

## Save semantics

- `UnlockedElements` is persisted as part of `GameState`.
- `InjectionElement` is persisted as part of `GameState` and normalized to Hydrogen if a save is missing it or points at a locked/unknown Element.
- `BestStats` is an additive persisted field on `GameState`.
- Legacy saves that predate either field load with safe defaults and do not require a save-version bump.

## Related

- `internal/render/periodic_table.go` — table rendering, hit-testing, focus card, unlock action.
- `internal/render/game.go` — Codex interaction routing.
- `internal/sim/economy.go` — catalog, purchase logic, research constants.
- `internal/sim/tick.go` — collection-time best-stat updates.
- `docs/features/value-formula.md` — how research factors into collected value.
