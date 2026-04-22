# Particle Accelerator At Home — Game Overview

This document is the single source of truth for the game's concept, terminology, and scope. Feature-level detail lives in `docs/features/`; architectural decisions live in `docs/adr/`. Update this doc when the game model changes; push implementation detail down into feature docs.

## Concept
A modular grid where **Subjects** (particles) are injected, routed through player-placed **Accelerator Components** that modify their properties, and collected at endpoints. Collection yields **$USD** and advances research on the Subject's **Element**. $USD + research buys upgrades that unlock heavier Elements, better Components, and (eventually) bigger grids.

Genre: incremental / idle. Platform: browser (Ebitengine → WASM → GitHub Pages), with desktop iteration via `go run ./cmd/game`.

## Terminology (canonical)
| Term | Meaning |
|---|---|
| $USD | In-game currency. |
| Subject | A particle currently in flight on the grid. |
| Element | Type of Subject (Hydrogen, Helium, ...). Research is tracked per Element. |
| Accelerator Component | Any placeable grid cell. |
| Mass | Physical weight of a Subject. Feeds into collected value. |
| Load | Grid-occupancy cost of a Subject. The accelerator has a **Max Load** cap. Injecting adds to used Load; collecting or losing a Subject frees it. |
| Tick | One logical simulation step. |

> The bare word "weight" is **ambiguous** and must not appear in code or UI — use **Mass** (physics) or **Load** (grid capacity) explicitly.

## Subject model
Every Subject carries:
- `Element`
- `Mass` — derived from Element plus any modifiers applied en route
- `Speed` — cells per tick
- `Magnetism`
- `Charge` (reserved for future)
- `Direction` — N/E/S/W
- `Position` — grid cell

Every Accelerator Component is conceptually a pure function `(Subject) → Subject`, optionally with a direction override. Adding a new Component means defining a new such function plus its sprite.

## Simulation model
- **Fixed logical tick rate**, user-configurable. Logical state advances only on ticks. This keeps the simulation deterministic for saves and offline progress. The constant lives at `sim.DefaultTickRate`.
  - **Target**: 60 Hz with render-side interpolation (Subjects lerp between `PrevPosition` and `Position` by the fractional tick).
  - **Current (Phase 1)**: 10 Hz with **tick-granular rendering** — Subjects snap to their logical cell, no interpolation. At 60 Hz + Speed=1, a Subject would cross the 5×5 grid in ~80 ms and teleport; interpolation is Phase 3 work, so we run the simulation slower meanwhile. Raise the default back to 60 once interpolation lands in `internal/render`.
- Multiple Subjects may be on-grid simultaneously, capped by **Max Load**.
- Collision handling (two Subjects in the same cell on the same tick) is TBD; MVP rule: ignore, both pass through.

## Accelerator Components (initial set)
- **Injector** — spawns a Subject every N ticks in its configured Direction. Blocks spawn when Max Load is reached.
- **Simple Accelerator** — `+1` Speed.
- **Mesh Grid** — `×0.5` Speed (rounded).
- **Magnetiser** — `+1` Magnetism.
- **Rotator** — redirects the Subject (configurable angle; 90° MVP).
- **Collector** — endpoint. Removes the Subject, awards $USD and Element research.

### Design principle — speed bands
Some Components should only trigger (or change behaviour) within specific Speed ranges. This is what makes Mesh Grid a *tool* instead of a trap. Exact bands per Component live in the relevant feature doc.

## Value formula (on collection)
Collected $USD is a function of: `Mass`, `Speed`, `Magnetism`, the Element's base multiplier, its per-Element research level, and global multipliers. The exact formula is TBD in a dedicated feature doc. Design constraint: each input axis should feel individually meaningful — no single axis dominating.

## Progression axes
1. **Per-Component tiers** — e.g. Simple Accelerator T1 → T3 (`+1` → `+3` Speed). Bought with $USD.
2. **Per-Element research** — collecting Subjects of an Element levels up its research, multiplying that Element's collected value. Research also gates heavier Elements.
3. **Global upgrades** — cross-cutting $USD sinks ("all Collectors +10%", "Injectors fire 2× as fast", etc.).
4. **Prestige layer (future)** — resets $USD and grid; awards a permanent meta-currency that buys, among other things, **Max Load** and **grid size** upgrades.

## Periodic Table (codex)
Dedicated screen listing every known Element with: research level, best stats (max Speed seen, max Mass, max collected value), unlock status, and cost to unlock the next.

## Grid
- Starts **5×5**.
- Upgradeable via prestige layer (design deferred).
- Rendering: each cell is a **two-layer sprite** (top + bottom of the accelerator tube). The Subject is z-ordered between the two layers, so it visually passes *through* each Component.

## Rendering / resolution
- **Logical resolution: 1280×720 (16:9)**. Ebitengine scales this to the window; assets are authored at 1×.
- Integer-scales cleanly to 1080p (1.5×) and 1440p (2×); downscales acceptably on 1366×768 laptops.
- Layout intent: roughly a 720×720 grid area (supports up to ~15×15 at 48 px cells, enough runway for phase-4 grid expansion) with a ~560 px UI column for economy, upgrades, and the codex.
- Fixed logical resolution with letterboxing for now; revisit responsive reflow if ultrawide/mobile support becomes a priority.

## Persistence
- **WASM**: LocalStorage.
- **Desktop**: file (already stubbed in `internal/save/`).
- Save contents: grid layout, $USD, per-Element research, unlocked upgrades, prestige state.
- **Offline progress**: on load, fast-forward the simulation by the elapsed wall-clock time. The deterministic tick model makes this tractable.

## Scope & phasing
MVP-first. Each phase ends with a playable build.

**Phase 1 — Core loop playable**
- Components: Injector, Simple Accelerator, Rotator, Collector
- One Element (Hydrogen)
- $USD economy with per-Component upgrades
- Save/load (no offline yet)
- 5×5 fixed grid

**Phase 2 — Depth**
- Components: Mesh Grid, Magnetiser
- Speed bands
- Second Element
- Per-Element research + Periodic Table screen

**Phase 3 — Polish**
- Global upgrades
- Offline progress
- More Elements
- Two-layer sprite rendering
- Render-side tick interpolation (lerp Subject position between ticks so the default tick rate can return to 60 Hz without visually teleporting)

**Phase 4 — Prestige**
- Reset layer
- Max Load upgrades
- Grid-size upgrades

## Open questions (resolve in feature docs)
- Exact collected-value formula.
- Collision behaviour when multiple Subjects share a cell.
- Injector tick rate — constant, or upgradeable per Injector?
- Speed-band boundaries per Component.
- Prestige-layer currency name and upgrade tree shape.

## Related docs
- `docs/features/` — one file per feature (e.g. `component-rotator.md`, `value-formula.md`).
- `docs/adr/` — architectural decisions (e.g. tick model, save format, rename from `orb`/`upgraders` to canonical terminology).
