# ADR 0012 — Component inventory modal

**Status:** accepted.
**Date:** 2026-04-24.

## Context

The right-side Components palette (`internal/render/palette.go`) was an always-visible vertical list of every placeable tool. Phase 3 brought it to 10 entries (Injector, Accelerator, Mesh Grid, Magnetiser, Resonator, Catalyst, Duplicator, Elbow, Collector, Erase). More are queued: prestige-layer components and future research-gated unlocks. An always-on list does not scale.

A modal picker also gives us a natural surface for **canonical component descriptions**, which up to now lived only in `docs/features/*.md` and were not reachable in-game.

Three shapes were considered:

1. **Keep the always-visible list, add a tooltip on hover.** Cheapest but does not solve the clutter — every new component still permanently consumes screen height. Tooltips also fight with the existing keyboard hint footer.
2. **Replace with a hotbar.** Common in incremental/builder games, but couples placement to a small fixed slot count and re-opens the question of how the player browses everything that isn't in the hotbar. Defer until we have a reason for the constraint.
3. **Modal picker, opened by `E` and a header button.** Frees the sidebar, gives plenty of room for full descriptions, matches the existing `Settings` / `Codex` modal pattern (`internal/render/settings.go`, `internal/render/periodic_table.go`).

We chose (3). The right sidebar is retained but shrunk to a single "Selected" indicator card plus an "Open Inventory" button — the player always sees what they are about to place, and the keybind affordance stays discoverable.

## Decision

**1. New UI state: `InventoryOpen`, `InventoryHovered`.**

```go
// internal/ui/state.go
type UIState struct {
    // ... existing fields
    InventoryOpen    bool
    InventoryHovered Tool
}
```

Both are session-scoped (not persisted), matching `SettingsOpen` / `CodexOpen`. `InventoryHovered` resets to `ToolNone` whenever the modal closes.

**2. Component copy lives in `internal/ui/component_info.go`.**

```go
type ToolInfo struct {
    Name        string
    Tagline     string
    Description string
}
var ToolInfoCatalog = map[Tool]ToolInfo{ ... }
```

Both the modal's description panel and the compact palette's "Selected" card read from this. `internal/ui/component_info_test.go` asserts every tool in `PlaceableTools` has an entry with non-empty `Name` and `Description`.

**Rejected alternatives:**
- *Attach description fields to `sim.ComponentCatalog`* — couples user-facing copy to the simulation package, which should remain headless. `sim` is the wrong layer for UI strings.
- *Parse `docs/features/*.md` at build time* — adds a toolchain step for marginal benefit. The feature docs target a different reader (designers, contributors) than the in-game tooltip (players). Allowing the two to diverge is fine.

**3. Lock predicate centralised in `internal/ui/lock.go`.**

```go
func IsToolUnlocked(s *sim.GameState, t Tool) bool
func ToolLockReason(s *sim.GameState, t Tool) string
func KindForTool(t Tool) sim.ComponentKind
var PlaceableTools = []Tool{ ... }
```

All tool lock checks defer to `IsToolUnlocked`. `KindForTool` replaces the old `palette.kindForTool` and `input.toolKind` duplicates. `PlaceableTools` is the canonical layout order for the modal grid, so reshuffling card order is a one-line edit.

This pulls `internal/sim` into `internal/ui`'s import graph for the first time. `internal/sim` does not import `internal/ui` (verified) so there is no cycle.

**4. Modal layout: 960×600, 4×3 card grid, right-side description panel.**

Card grid sized to fit comfortably above the planned tool count for at least the next few content drops (12 slots, 11 used today). Description panel is docked on the right rather than floating because:

- Hit-testing is simpler — no edge-clamping when hovering a card near the modal's edge.
- The panel is always visible, so the player can scan it without precise mouse aim.
- Matches the Codex card-on-the-right pattern (`periodic_table.go`).

Hover behaviour: when the cursor leaves a card the description panel keeps showing the last hovered tool until the cursor enters another card or the modal closes. Clearing on every blank pixel felt flickery in playtest.

**5. Click semantics.**

| Click target | Effect |
|---|---|
| Unlocked card | `Selected = tool`, modal closes. |
| Locked card | No-op. |
| Close button | Modal closes; selection unchanged. |
| Outside modal panel | Modal closes; selection unchanged. |
| Selected card | Re-selects (no toggle-off). The old palette toggled selection on second click; the inventory does not, because the modal closes on the click and toggle-off has no clear UX win when the player has to reopen the modal anyway. To deselect, the player can press `E` and click the same tool again — or place into an empty cell (placement keeps the selection per the cost-flow design). |

**6. Selection persistence is unchanged.**

`input.PlaceFromTool` (`internal/input/input.go`) already leaves `u.Selected` untouched after a successful or failed placement, and auto-purchases the next unit when stock empties if affordable. This satisfies "stays selected after placing one (assuming they have stock)" with zero new code. The feature doc calls this out explicitly so it doesn't get accidentally regressed.

**7. The `E` key opens / closes the modal.**

`E` is unbound today. Arbitrary letter keys are reserved for global toggles (we have `T` for trails). `E` was preferred over `I` because `I` may collide with future "Information" affordances and because `E` is geographically friendly on QWERTY for left-handed mouse use. Pressing `E` while another modal is open closes that modal first, so the player is never stuck with two overlays.

**8. Right-side panel shrinks; doesn't disappear.**

Removing the panel entirely was considered. Rejected because:

- The selected-tool indicator gives at-a-glance feedback during multi-placement workflows. Without it, the player has to re-open the modal to verify what they're holding.
- The "Open Inventory (E)" button surfaces the keybind for new players who haven't read any docs.
- The grid would have to expand into the freed space, which means changing `cellSize`, the layout maths, and the sprite scaling — disproportionate scope for this change.

If we later decide a wider grid is the better trade, deleting `drawPalette`/`drawSelectedCard` and reflowing `layout.go` is a small follow-up.

## Consequences

**Wins**
- One new modal file plus a small UI-state delta. Existing patterns (settings, codex) provide all the rendering primitives.
- Picker scales to many more components without consuming permanent screen height.
- First in-game home for canonical component copy — a hook for future onboarding flows.
- Shared `IsToolUnlocked` / `KindForTool` removes three previous duplications.

**Costs**
- Two surfaces now show the selected tool (the modal's selected highlight + the sidebar's "Selected" card). They could drift if a future change touches one and not the other. Mitigation: the sidebar reads from `ToolInfoCatalog` and `KindForTool`, so most copy/data changes propagate automatically.
- The modal currently caps at 12 cards (4×3). At 13+ tools we will need to either grow the modal, reduce card size, or paginate. Easy to refactor when it happens, but keep it in mind when adding new components.
- `internal/ui` now imports `internal/sim`. Acceptable — no cycle today, and the lock predicate genuinely needs sim state — but it raises the bar for keeping `sim` headless.

## Related

- `internal/render/inventory.go` (new) — modal renderer, click handler, hover handler, layout helpers.
- `internal/render/inventory_test.go` (new) — card layout, click, close, outside-click tests.
- `internal/ui/component_info.go` (new) — canonical copy.
- `internal/ui/component_info_test.go` (new) — coverage assertion across `PlaceableTools`.
- `internal/ui/lock.go` (new) — `IsToolUnlocked`, `ToolLockReason`, `KindForTool`, `PlaceableTools`.
- `internal/ui/state.go` — `InventoryOpen`, `InventoryHovered` fields.
- `internal/render/header.go` — "Inventory (E)" button.
- `internal/render/game.go` — `E` key handler, modal-priority guard, header click, Draw call.
- `internal/render/palette.go` — shrunk to "Selected" card + Open Inventory button.
- `internal/render/draw_util.go` — `drawTextWrapped`, `wrappedHeight` helpers added for the description panel.
- `internal/input/input.go` — `toolKind` removed, replaced by shared `ui.KindForTool`.
- `docs/features/0014-inventory.md` — player-facing feature doc.
- ADR 0005 — component cost / inventory accounting that the modal surfaces.
- ADR 0007 — Codex modal pattern this design follows.
- ADR 0011 — tier system; tier upgrades remain in their own UI surface (not the inventory).
