# ADR 0013 — Helper notifications and history

**Status:** accepted.
**Date:** 2026-04-24.

## Context

The game needs lightweight onboarding and contextual explanations without forcing the player into external docs. Two near-term use cases drove the design:

1. A milestone helper when the player's current $USD balance first reaches `$5`, teaching that `E` opens the Inventory.
2. A contextual helper when the player hovers a placed Accelerator Component and presses `H`, showing component copy and live stats.

The existing UI has several blocking modal surfaces (`Settings`, `Inventory`, `Codex`, `Collection Log`) and a persisted `CollectionLog` for recent payouts. There was no generic helper surface or notification history.

## Decision

**1. Helper modal state is session-scoped UI state.**

`internal/ui.UIState` owns whether the helper is open, its header/body, and its anchor mode/position. This matches the other render-only modal flags and keeps the simulation package headless.

**2. Logged notification history is persisted on `GameState`.**

`GameState.NotificationLog []NotificationEntry` stores logged helper notifications, newest first, capped to `sim.MaxNotificationLogEntries` (`50`). Entries contain header, body, local `HH:MM` display time, and creation tick.

This is additive save data and does not require an envelope version bump. Old saves load with an empty history.

**3. One-shot milestone state is persisted on `GameState`.**

`GameState.ShownHelperMilestones map[string]bool` records helper IDs that have already displayed. This prevents milestone helpers from reappearing every session after their trigger condition is already true. Hard reset naturally clears the map by replacing state with `NewGameState()`.

**4. Callers explicitly choose whether to log.**

Milestone/tutorial helpers log to history. Contextual `H` helpers do not, because repeated hover-help would quickly pollute the history with low-value entries.

**5. Helpers are blocking overlays, not passive tooltips.**

While a helper is open, input behind it is swallowed. `Escape` and the Close button dismiss it. The simulation continues ticking; the decision only blocks interaction.

Blocking was chosen over non-blocking popovers because tutorial messages should not be missed or accidentally clicked through, and it fits the existing modal input structure in `internal/render/game.go`.

**6. Cursor helpers flip and clamp in logical screen space.**

Cursor-anchored helpers spawn to the right/below the cursor on the left/top halves of the screen, and to the left/above on the right/bottom halves. Final bounds are clamped to the logical `1280x720` screen.

This is simpler than maintaining per-edge arrow direction assets and is enough to prevent off-screen content with the current fixed logical layout.

## Consequences

**Wins**

- One reusable helper modal supports both tutorials and contextual component information.
- Persisted history makes milestone messages reviewable after dismissal.
- One-shot milestone IDs prevent repeated onboarding spam on reload.
- `H` help reuses `ToolInfoCatalog`, avoiding a second copy source for component descriptions.

**Costs**

- `GameState` now includes user-facing notification history, not just simulation/economy data. This is acceptable because saves are already the source of persisted player-facing logs (`CollectionLog`).
- History currently stores display-time text (`HH:MM`) rather than a full timestamp. If future features need sorting across days or elapsed-time calculations, add an ISO timestamp field while preserving `time_hhmm` for old entries.
- Notification History uses wheel scrolling, but no scrollbar thumb yet. Add one when the visual language for long lists is settled.

## Alternatives Considered

- **Session-only history.** Rejected because Settings history should remain useful after browser reloads.
- **Always log every helper.** Rejected because `H` component help is repeatable and would drown out tutorial notifications.
- **Header tab for notifications.** Rejected for now; Settings is the right home until notification history becomes a core loop surface.
- **Pause the simulation while helpers are open.** Rejected because the fixed-tick simulation can safely continue, and pausing an idle game for tutorials would make timing/cooldown behavior feel inconsistent.

## Related

- `docs/features/0017-helper-notifications.md`.
- `internal/sim/notification.go`.
- `internal/render/helper_system.go`.
- `internal/render/helper_modal.go`.
- `internal/render/notification_history.go`.
- `internal/render/settings.go`.
- `internal/render/game.go`.
