# Periodic Table (Codex)

**Status:** Phase 2.

A modal screen reached from the "Codex" button in the game header. Lists every Element defined in `sim.ElementCatalog` and shows the player's progress on each.

## Columns

| Column     | Content                                                                                          |
|------------|--------------------------------------------------------------------------------------------------|
| Element    | `{Symbol}  {Name}`.                                                                              |
| Research   | `GameState.Research[e]` — subjects collected of that Element.                                    |
| Multiplier | `baseMultiplier × (1 + Research / ResearchK)` — the current effective multiplier, updated live.  |
| Status     | One of: `Unlocked`, `Unlock for $N` (purchasable), `Locked · N more H research` (research-gated). |

## Unlock flow

Each Element's row has an action state driven by two predicates:

1. **Research gate** — `Research[UnlocksFrom] >= ResearchThreshold`. While below, the row is `Locked` and shows how much research is still needed.
2. **Purchase gate** — once the research gate is clear, the row shows an `Unlock for $N` button. Clicking deducts `UnlockCost` from `USD` and flips `GameState.UnlockedElements[e] = true`.

Unlocking an Element activates the corresponding Injector entry in the Component palette.

Current catalog (`internal/sim/economy.go`):

| Element  | UnlocksFrom | ResearchThreshold | UnlockCost |
|----------|-------------|-------------------|------------|
| Hydrogen | —           | 0 (seeded unlocked) | 0        |
| Helium   | Hydrogen    | 10                | 500        |

## Save semantics

`UnlockedElements` is persisted as part of `GameState`. Legacy saves that predate the field load with `{Hydrogen: true}` via the nil-guard in `Load()` — no save-version bump.

## Related

- `internal/render/periodic_table.go` — modal rendering + click handling.
- `internal/sim/economy.go` — catalog, purchase logic, research constants.
- `docs/features/value-formula.md` — how research factors into collected value.
