# Inventory

**Status:** Phase 3.

## Concept

The Inventory is a full-screen modal that holds every placeable Accelerator Component. It replaces the older always-visible right-side palette as the canonical picker. The right-side panel still exists but shrinks to a single "Selected" indicator card plus an "Open Inventory" button, so the player always knows what they're about to place without sacrificing screen real-estate to the full list.

## Opening and closing

- **`E` key** toggles the modal. Closes any other modal that is already open.
- **Header button "Inventory (E)"** (left of Codex) opens it.
- **Close button** (top-right of the modal) dismisses it.
- Clicking outside the modal panel also dismisses it.

## Layout

The modal is a `960 × 600` panel centred on the logical 1280×720 screen.

- **Card grid** on the left: 4 columns × 3 rows of `140 × 140` cards. Each `ui.PlaceableTool` gets one card; the order in `internal/ui/lock.go` (`PlaceableTools`) determines layout order.
- **Description panel** on the right (`280 px` wide): canonical name, tagline, full description, and a stat strip (Owned, Placed, Available, Next purchase cost) for whichever card the cursor is currently over.

### Card anatomy

| Region | Content |
|---|---|
| Top strip | Available count `xN` (or `—` when locked). |
| Middle | Component icon (sprite) at `80×80`. Falls back to a colour swatch if no sprite is wired. |
| Bottom strip | Next purchase cost (`$1.2K`). Cost text turns red when the player is empty AND can't afford another. Locked cards show `Locked` instead. |
| Hover border | The hovered card's border switches to `colorText` for a clearer focus ring. |
| Selected border | The currently-selected tool's card uses `colorSelected` for its background. |
| Lock dim | Locked or empty-and-unaffordable cards get a translucent overlay over the icon. |

### Description panel states

- **Nothing hovered:** "Hover a component for details."
- **Hovered:** name (title font), tagline (highlight colour), description (body), then the stat strip pinned to the bottom of the panel.
- **Locked:** an additional red "Locked: …" line explains why (sourced from `ui.ToolLockReason`).

## Selecting a component

Clicking an unlocked card sets `UIState.Selected` and closes the modal. The selection persists across placements: `input.PlaceFromTool` does not clear `Selected`, and auto-purchases the next unit when stock runs out (provided the player can afford it). The player can place several of the same component without re-opening the modal.

Clicking a locked card is a no-op — the description panel already explains why it can't be selected.

## Locking

No inventory tools are Element-specific. There is one Injector card; its emitted Element is selected globally from unlocked Elements in the Codex. The lock check is centralised in `ui.IsToolUnlocked` for future gated tools so the inventory, the compact palette, and any future surface (hotbar, achievements, etc.) all read from one source.

## Right-side panel after the rewrite

The compact panel shows:

- Header: "Selected".
- A card with the currently-selected tool's icon, name, tagline, available count, and next cost. Empty state: "Nothing selected · Press E to open inventory".
- An "Open Inventory (E)" button.
- The unchanged keybind footer ("Left-click: place / reconfigure" etc).

## Related

- `internal/render/inventory.go` — modal draw + click + hover logic.
- `internal/ui/component_info.go` — single source of truth for component name/tagline/description.
- `internal/ui/lock.go` — lock predicate, lock reason, tool→kind mapping, and the canonical `PlaceableTools` order.
- `docs/adr/0012-component-inventory-modal.md` — design decisions behind this feature.
- `docs/features/component-cost.md` — purchase cost curve consumed by the cost label.
- `docs/features/component-tiers.md` — orthogonal progression axis surfaced in card stats.
