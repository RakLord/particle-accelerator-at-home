# Helper Notifications

**Status:** Phase 3.

Helper notifications are blocking in-game tutorial/info modals. They are used for milestone onboarding and on-demand component help.

## Modal Shape

- Header text.
- Wrapped paragraph/body text.
- Close button.
- `Escape` also closes the topmost helper modal.

Helpers can spawn in two modes:

- **Centered:** used for milestones and tutorial beats.
- **Cursor anchored:** used for contextual help, such as pressing `H` over a placed Accelerator Component. The modal flips left/right and above/below based on cursor side, then clamps to the logical `1280x720` screen so it does not run off-screen.

While open, a helper blocks gameplay and other UI input behind it. The simulation continues ticking; only input is swallowed.

## Milestones

Milestone helpers are one-shot per save. `GameState.ShownHelperMilestones` stores string IDs that have already been displayed. Hard reset clears the map by restoring a fresh `GameState`.

Current milestone:

| ID | Trigger | Header | Body |
|---|---|---|---|
| `first-five-usd` | Current $USD balance reaches at least `$5` | `Inventory Available` | `You can press E to open the Inventory. In the Inventory you can buy new components for your Accelerator.` |

## Component Help

Pressing `H` while hovering a placed component/Collector opens a cursor-anchored helper. This is only active during normal grid interaction; it does not fire while Settings, Inventory, Codex, Collection Log, Notification History, or another helper modal is open.

Component help reuses `internal/ui/component_info.go` copy and adds live stats where applicable:

- Tier.
- Owned.
- Placed.
- Available.
- Next component purchase cost.

Component help is intentionally **not logged** to notification history.

## Notification History

Logged helper notifications are persisted on `GameState.NotificationLog`, newest first, capped to 50 entries. Each entry stores:

- Header.
- Body.
- Local display timestamp in `HH:MM` format.
- Tick number at creation.

Settings includes a `Notification History` button that opens a larger submodal. The history view shows logged helpers newest-first and supports mouse-wheel scrolling when more entries exist than fit on screen.

## Logging Rules

- Milestone/tutorial helpers log by default.
- Contextual `H` component helpers do not log.
- Future helper callers must choose whether to log each notification.

## Related

- `internal/sim/notification.go` — persisted notification log and milestone helpers.
- `internal/render/helper_system.go` — milestone checks and helper construction.
- `internal/render/helper_modal.go` — blocking helper modal layout/draw.
- `internal/render/notification_history.go` — Settings history submodal.
- `internal/ui/state.go` — session-scoped helper/history open state.
- ADR 0013 — save/UI decision record.
