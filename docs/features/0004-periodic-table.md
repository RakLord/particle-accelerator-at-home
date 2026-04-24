# Periodic Table (Codex)

**Status:** Phase 2.

The Codex is a large overlay reached from the "Codex" button in the header. It presents Elements in a visually recognizable periodic-table layout, then shows detail in a centered stat card for the currently focused Element.

## Layout

The table uses real periodic-table coordinates, not a compact list.

- Each Element in `sim.ElementCatalog` supplies display metadata: `AtomicNumber`, `Period`, and `Group`.
- The Codex renders an 18-column table frame so early-game Elements can sit in their canonical positions even when the rest of the table is still empty.
- Current shipped positions:

| Element | AtomicNumber | Period | Group |
|---|---:|---:|---:|
| Hydrogen | 1 | 1 | 1 |
| Helium | 2 | 1 | 18 |
| Lithium | 3 | 2 | 1 |
| Beryllium | 4 | 2 | 2 |
| Boron | 5 | 2 | 13 |
| Carbon | 6 | 2 | 14 |
| Nitrogen | 7 | 2 | 15 |
| Oxygen | 8 | 2 | 16 |
| Fluorine | 9 | 2 | 17 |
| Neon | 10 | 2 | 18 |
| Sodium | 11 | 3 | 1 |
| Magnesium | 12 | 3 | 2 |
| Aluminium | 13 | 3 | 13 |
| Silicon | 14 | 3 | 14 |
| Phosphorus | 15 | 3 | 15 |
| Sulfur | 16 | 3 | 16 |
| Chlorine | 17 | 3 | 17 |
| Argon | 18 | 3 | 18 |
| Potassium | 19 | 4 | 1 |
| Calcium | 20 | 4 | 2 |

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
- base Mass used when injecting the Element
- base Speed used when injecting the Element
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

| Element | BaseMass | BaseSpeed | Multiplier | UnlocksFrom | ResearchThreshold | UnlockCost |
|---|---:|---:|---:|---|---:|---:|
| Hydrogen | 1.008 | 2 | 1.0 | - | 0 | 0 |
| Helium | 4.003 | 2 | 1.5 | Hydrogen | 10 | 500 |
| Lithium | 6.94 | 1 | 1.8 | Helium | 12 | 2,000 |
| Beryllium | 9.012 | 1 | 2.1 | Lithium | 14 | 8,000 |
| Boron | 10.81 | 1 | 2.4 | Beryllium | 16 | 30,000 |
| Carbon | 12.011 | 1 | 2.8 | Boron | 18 | 100,000 |
| Nitrogen | 14.007 | 1 | 3.2 | Carbon | 20 | 300,000 |
| Oxygen | 15.999 | 1 | 3.6 | Nitrogen | 22 | 900,000 |
| Fluorine | 18.998 | 1 | 3.8 | Oxygen | 24 | 2,500,000 |
| Neon | 20.180 | 1 | 4.0 | Fluorine | 26 | 7,500,000 |
| Sodium | 22.990 | 1 | 4.5 | Neon | 28 | 20,000,000 |
| Magnesium | 24.305 | 1 | 5.0 | Sodium | 30 | 55,000,000 |
| Aluminium | 26.982 | 1 | 5.7 | Magnesium | 32 | 150,000,000 |
| Silicon | 28.085 | 1 | 6.5 | Aluminium | 34 | 400,000,000 |
| Phosphorus | 30.974 | 1 | 7.2 | Silicon | 36 | 1,000,000,000 |
| Sulfur | 32.06 | 1 | 8.0 | Phosphorus | 38 | 2,500,000,000 |
| Chlorine | 35.45 | 1 | 8.6 | Sulfur | 40 | 6,000,000,000 |
| Argon | 39.948 | 1 | 9.0 | Chlorine | 42 | 15,000,000,000 |
| Potassium | 39.098 | 1 | 10.0 | Argon | 45 | 40,000,000,000 |
| Calcium | 40.078 | 1 | 12.0 | Potassium | 50 | 100,000,000,000 |

`BaseMass` uses standard atomic mass values so heavier Elements naturally pay more through the value formula. `BaseSpeed` is intentionally coarse: Hydrogen and Helium start at Speed 2, while heavier Elements start at Speed 1 until a future charge/energy model can represent acceleration more directly.

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
- `docs/features/0001-value-formula.md` — how research factors into collected value.
