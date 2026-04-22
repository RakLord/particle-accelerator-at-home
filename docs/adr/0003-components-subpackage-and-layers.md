# ADR 0003 — Components subpackage + Layer seam

**Status:** accepted.
**Date:** 2026-04-22.

## Context

By the end of Phase 2 the `internal/sim` package held five `component_*.go` files plus a tick loop that type-asserted directly to `*Injector`. Roadmap calls for more components each phase and multiple reset layers (Genesis today; unnamed layers to follow). Two pain points were foreseeable:

1. Adding a new Component touched a central switch in `save.newComponentByKind` and, for source-type components, `tick.injectorSpawns`.
2. Everything lived in one flat package, so there was nowhere obvious to hang per-layer scoping when the next reset layer arrives.

## Decision

**1. Extract concrete Components into `internal/sim/components`.**
- `sim` retains the `Component` interface, `ComponentKind`, the registry (`RegisterComponent` / `newComponentByKind`), and a new capability interface `Spawner` that source-type Components implement.
- Each concrete Component lives in one file under `internal/sim/components/`, owns its `KindFoo` constant, and calls `sim.RegisterComponent` from `init()`.
- The binary (`cmd/game/main.go`) blank-imports `internal/sim/components` to trigger registration. The render and input packages import it directly for the types they switch on.

**2. Introduce `sim.Layer` and `LayerGenesis`.**
- `GameState.Layer` seeded to `LayerGenesis` on `NewGameState`. Additive save field; nil-guard in `Load` defaults legacy saves to Genesis. No save-version bump.
- No per-layer subfolders under `components/` yet — flat is fine until there's a second layer with a concrete need.

**3. Scope of what *didn't* move.**
- `Subject`, `Direction`, `Element`, `Position`, `Grid`, `Cell`, `GameState` stay in `sim`. They're the vocabulary every Component speaks.
- Render-side UI metadata (swatch color, glyph drawing) stays in `internal/render` — moving it per-component would force ebiten into the components package.

## Consequences

**Wins**
- Adding a Component: drop one file in `components/`. No edits to `tick.go`, `save.go`, or the registry.
- Save-format stability: `ComponentKind` strings are the on-disk identifiers. Moving the constants between packages doesn't change them.
- Layer concept exists as a seam without committing to a shape.

**Costs**
- Tests that use concrete Components live in `components_test` (external package) or under `components/`. The in-package `sim` test file lost the cell round-trip — it's duplicated in spirit by `components/save_test.go`.
- `input` and `render` now import both `sim` and `sim/components`. Acceptable; no import cycle.
- Spawner is a capability interface; anyone iterating cells for spawn-like behaviour has to remember to check for it. Documented in `components.go`.

## Alternatives considered

- **Keep concretes in `sim` but rely on registry.** Avoids import-path churn but doesn't give future layer/grouping structure a home. Rejected — the user explicitly asked for the subpackage.
- **Per-layer subpackages (`components/genesis/`, etc.).** Premature until the second layer has gameplay. Revisit when `LayerX` is named.
- **Render-side self-registration** (each component registers its swatch/glyph). Rejected for now: it couples the components package to ebiten, which would drag in a graphics dependency for headless tests.

## Related

- `internal/sim/components.go` — interface, registry, Spawner.
- `internal/sim/components/` — concrete types.
- `internal/sim/layer.go` — Layer type + LayerGenesis.
- ADR-0002 — save schema versioning (still v1; this change is additive only).
