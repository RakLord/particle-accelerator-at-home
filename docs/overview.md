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

Every Accelerator Component is conceptually a pure function `(Subject, context) → Subject`, optionally with a direction override. The context (grid read, per-Element research, tick, global modifiers, tier level) is covered by `docs/adr/0008-apply-context-and-grid-view.md`. Components that emit additional Subjects (e.g. Duplicator) implement a sibling capability interface — see `docs/adr/0009-subject-emitter-capability.md`. Adding a new Component means defining its function plus its sprite.

When adding or retuning Components, use `docs/features/component-creation-and-balancing.md` for the implementation checklist and purchase-cost balancing workflow.

## Simulation model
- **Fixed logical tick rate**, user-configurable. Logical state advances only on ticks. This keeps the simulation deterministic for saves and offline progress. The constant lives at `sim.DefaultTickRate`.
  - Render-side interpolation is live (see `docs/features/smooth-motion.md`): Subjects glide between ticks along a recorded per-tick `Path`, with quarter arcs through rotator cells. A `sim.SpeedDivisor` of 10 means base `Speed=1` traverses one cell every 10 ticks; the tick rate itself stays at 10 Hz for now.
- Multiple Subjects may be on-grid simultaneously, capped by **Max Load**.
- Collision handling (two Subjects in the same cell on the same tick) is TBD; MVP rule: ignore, both pass through.

## Accelerator Components (initial set)
- **Injector** — emits the Codex-selected Element as a Subject when the player presses Inject. The global Inject action has a 5-second base cooldown and blocks when Max Load is reached.
- **Simple Accelerator** — `+1` Speed at T1 (see `docs/features/component-tiers.md`).
- **Mesh Grid** — `÷2` Speed at T1.
- **Magnetiser** — `+1` Magnetism at T1.
- **Rotator** — redirects the Subject (configurable angle; 90° MVP).
- **Collector** — endpoint. Removes the Subject, awards $USD and Element research.

### Phase-3 additions
- **Resonator** — neighbour-aware Speed boost (`docs/features/component-resonator.md`).
- **Catalyst** — research-gated Mass multiplier (`docs/features/component-catalyst.md`).
- **Duplicator** — T-junction that emits two Subjects (`docs/features/component-duplicator.md`).

### Design principle — speed bands
Some Components should only trigger (or change behaviour) within specific Speed ranges. This is what makes Mesh Grid a *tool* instead of a trap. Exact bands per Component live in the relevant feature doc.

## Value formula (on collection)
Collected $USD is a function of: `Mass`, `Speed`, `Magnetism`, the Element's base multiplier, its per-Element research level, and global multipliers. The exact formula is TBD in a dedicated feature doc. Design constraint: each input axis should feel individually meaningful — no single axis dominating.

## Progression axes
1. **Per-Component tiers** — e.g. Simple Accelerator T1 → T3 (`+1` → `+3` Speed). Bought with $USD.
2. **Per-Element research** — collecting Subjects of an Element levels up its research, multiplying that Element's collected value. Research also gates heavier Elements.
3. **Global upgrades** — cross-cutting $USD sinks ("all Collectors +10%", "Inject cooldown 2× as fast", etc.).
4. **Reset layers (future)** — the game has multiple nested prestige layers. The base layer is **Genesis** (the game as shipped today); ascending to the next layer resets Genesis and awards a meta-currency. Each layer has its own Elements, Components, and currency context; meta-currency carries across. Layer names beyond Genesis are TBD. Represented in code as `sim.Layer` with `sim.LayerGenesis` seeded on `NewGameState`.

## Periodic Table (codex)
Dedicated screen styled as a real periodic table. Hovering or selecting an Element opens a centered stat card showing research level, best stats (max Speed, max Mass, max collected value), unlock status, cost to unlock the next, and the injection Element selector.

## Collection Log
Header modal showing the 10 most recent collected Subjects with the stats that fed their value: Element, Mass, Speed, Magnetism, research level, and awarded $USD.

## Load Bar / Injection
The grid area shows `Load: Current/Max` as a progress bar above the accelerator. The right-side panel has a manual Inject button. Pressing it commands every placed Injector to emit once, subject to the global injection cooldown and Max Load cap.

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

**Phase 1 — Core loop playable (completed)**
- Components: Injector, Simple Accelerator, Rotator, Collector
- One Element (Hydrogen)
- $USD economy with per-Component upgrades
- Save/load (no offline yet)
- 5×5 fixed grid

**Phase 2 — Depth (completed)**
- Components: Mesh Grid, Magnetiser
- Speed bands
- Second Element (Helium) with research-gated, $USD-purchasable unlock
- Per-Element research multiplier + Periodic Table (Codex) screen

**Phase 3 — Polish**
- Component tier primitive (`docs/features/component-tiers.md`).
- Global upgrades (`docs/features/global-upgrades.md`).
- New components: Resonator, Catalyst, Duplicator.
- Offline progress.
- More Elements.
- ~~Two-layer sprite rendering~~ — live.
- ~~Render-side tick interpolation~~ — live (`docs/features/smooth-motion.md`). Raising `DefaultTickRate` back to 60 is now a gameplay decision, not a rendering blocker.

**Phase 4 — Prestige**
- Reset layer
- Max Load upgrades
- Grid-size upgrades

## Open questions (resolve in feature docs)
- Exact collected-value formula.
- Collision behaviour when multiple Subjects share a cell.
- Injection cooldown upgrades and future auto-injection upgrade shape.
- Speed-band boundaries per Component.
- Prestige-layer currency name and upgrade tree shape.

## Related docs
- `docs/features/` — one file per feature (e.g. `component-rotator.md`, `value-formula.md`).
- `docs/adr/` — architectural decisions (e.g. tick model, save format, rename from `orb`/`upgraders` to canonical terminology).
