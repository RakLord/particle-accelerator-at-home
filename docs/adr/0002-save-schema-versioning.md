# ADR-0002: Versioned save schema

**Status:** Accepted
**Date:** 2026-04-22

## Context

The game needs to persist `GameState` (grid layout, $USD, per-Element research, upgrades, eventual prestige state) across sessions on both WASM (LocalStorage) and desktop (JSON file). The schema will change as phases 2–4 add Components, Elements, and upgrade axes. Retrofitting a version field after players have saves is painful — either you migrate blind, or you break existing saves.

## Decision

All saves are written through a versioned envelope:

```json
{
  "version": 1,
  "state": { ... full GameState ... }
}
```

- `version` is a monotonically increasing integer.
- Load reads `version` first and dispatches to a migration path if needed. Phase 1 handles only `version == 1`; unknown versions return an error and the game boots with default state.
- Bump `version` whenever the shape of `state` changes in a non-additive way. Purely additive changes (new optional field) do not require a version bump if default-zero-value semantics are acceptable.

The envelope is stored under a single key (`"state"`) via `internal/save.Write` / `internal/save.Read`. The platform split (`localstorage_js.go` / `file_desktop.go`) is unaffected.

## Rationale

- Versioning from v1 means migrations are a solved problem the first time we need one.
- A single key keeps the save atomic on both platforms; splitting fields across keys invites partial-write corruption.
- Unknown-version → default state is safer than attempting a best-effort migration on load.

## Consequences

- Every future save-shape change requires a decision: additive (no bump) or breaking (bump + migration).
- Tests for save/load must round-trip through the envelope, not just marshal `GameState` directly.
