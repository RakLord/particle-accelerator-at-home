# ADR-0001: Rename `Orb`/`Upgrader` to `Subject`/`Component`

**Status:** Accepted
**Date:** 2026-04-22

## Context

The scaffolding in `internal/sim/` uses the names `Orb` and `Upgrader`. `docs/overview.md` establishes the canonical game terminology as **Subject** (a particle in flight) and **Accelerator Component** (any placeable grid cell). `CLAUDE.md` flags these as load-bearing in code and notes that a rename is pending.

## Decision

Rename both types and their files before any further work lands:

- `internal/sim/orb.go` → `internal/sim/subject.go`
- `internal/sim/upgraders.go` → `internal/sim/components.go`
- type `Orb` → type `Subject`
- type `Upgrader` → type `Component`

New component implementations live in `internal/sim/component_*.go` (one file per concrete type).

## Rationale

- Scaffolding is ~50 lines of Go with no external consumers; renaming now is cheap and avoids accumulating uses of the legacy names.
- Keeping the legacy names creates a persistent mismatch between docs and code, which is worse than the one-time rename cost.
- File renames (not just type renames) are required so that `grep Subject` / `grep Component` returns the right results.

## Consequences

- All subsequent code and docs use `Subject` and `Component`.
- The word **"weight"** remains banned per `docs/overview.md`; use `Mass` or `Load`.
