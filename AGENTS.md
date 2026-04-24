# AGENTS.md

This file provides guidance to Codex, OpenCode, and other coding agents when working with code in this repository.

## What this is

An incremental/idle game built with **Ebitengine** in Go, targeting the browser via WASM (deployed to GitHub Pages) with desktop as the iteration target. The full game concept, terminology, and phasing lives in `docs/overview.md` -- **read it before making non-trivial changes**; it is the canonical source of truth for the game model. Feature-level detail belongs in `docs/features/`, architectural decisions in `docs/adr/`.

## Commands

```bash
# Desktop iteration (primary dev loop)
go run ./cmd/game

# WASM build (what CI does for GitHub Pages)
GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o web/game.wasm ./cmd/game
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/wasm_exec.js

# Standard Go tooling
go build ./...
go vet ./...
go test ./...                 # no tests yet; scaffold exists
go test ./internal/sim -run TestFoo   # run a single test once tests are added
```

CI (`.github/workflows/deploy.yml`) builds WASM on push to `master` and publishes `web/` to GitHub Pages. `web/game.wasm` and `web/wasm_exec.js` are gitignored -- they are build artifacts.

## Architecture

Entry point: `cmd/game/main.go` wires a `sim.Grid` into a `render.Game` (which implements `ebiten.Game`) and runs Ebitengine.

Packages under `internal/` are split by concern so the simulation can stay headless and deterministic:

- `internal/sim` -- grid, subjects-in-flight, ticks, and component behaviour. **No Ebitengine imports.** All game logic lives here.
- `internal/render` -- the `ebiten.Game` implementation. Reads `sim` state; never mutates it outside `Update()`'s call to `Tick()`.
- `internal/save` -- persistence with build-tagged implementations: `localstorage_js.go` (`//go:build js && wasm`) uses `syscall/js` -> `localStorage`; `file_desktop.go` (`//go:build !(js && wasm)`) writes JSON under `os.UserConfigDir()`. Keep this split -- do not collapse into a single file.
- `internal/input`, `internal/ui`, `internal/shader` -- stubs reserved for upcoming phases.

### Tick model (load-bearing invariant)

Logical state advances **only on fixed-rate ticks** (default 60 Hz). Rendering uses delta time for interpolation only. This is what makes saves and offline progress (fast-forwarding elapsed wall-clock time on load) tractable. Don't put gameplay state changes in `Draw` or make `Update` wall-clock-dependent.

### Component model

Every Accelerator Component is conceptually `(Subject) -> Subject`, optionally with a direction override. Adding a new component = one pure function + its sprite. The `sim.Upgrader` interface is the current placeholder for this.

Before adding a new Accelerator Component or changing component purchase costs, read `docs/features/component-creation-and-balancing.md`. It contains the implementation checklist, cost formula, soft-cap guidance, and balancing workflow that should stay consistent across components.

### Terminology

`docs/overview.md` defines canonical terms. Two are load-bearing in code:

- **Subject** = a particle in flight. **Accelerator Component** = any placeable cell. The scaffolding still uses the legacy names `Orb` (`internal/sim/orb.go`) and `Upgrader` (`internal/sim/upgraders.go`); an ADR-tracked rename to `Subject` / `Component` is pending. Prefer the canonical names in new code and docs; don't introduce *new* uses of `orb`/`upgrader`.
- The bare word **"weight"** is banned -- it's ambiguous. Use **Mass** (physics property of a Subject) or **Load** (grid-occupancy cost, capped by Max Load) explicitly.

### Rendering resolution

Design target is **logical 1280x720** with Ebitengine scaling to the window (see `docs/overview.md` for rationale). The current `render.Game.Layout` returns a square sized from `GridSize` -- this is scaffolding and will change when the UI column lands; don't treat the current layout as the intended one.
