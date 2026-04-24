# Prestige — Genesis Ascension

**Status:** Phase 4.

## Concept

Prestige is the game's first reset layer. The base layer (the game as shipped today) is **Genesis**; ascending out of Genesis resets most state and awards a **Bond** — a permanent compound bonus that persists across all future resets. Bonds are the meta-currency unlock; **Bond Points** earned alongside them feed the **Laboratory** upgrade tree.

Carbon (Z=6) is the prestige gate Element. Until the player has unlocked Carbon, placed a Binder, banked Subjects, crystallised Tokens, and synthesised at least one compound, the Prestige button stays hidden.

The companion features describe the parts:

- `docs/features/0022-component-binder.md` — the new Binder Component that banks Subjects.
- `docs/features/0021-bonds-and-tokens.md` — Token crystallisation and the compound (Bond) catalog.
- `docs/features/0023-laboratory.md` — the post-prestige Bond Point tree.
- `docs/features/0020-auto-injection.md` — the idle-injection mechanic unlocked by the Benzene Bond.

## Trigger

The Prestige button appears in the right-side panel once `len(GameState.BondsState) >= 1`. There is no $USD or research gate beyond synthesising a compound — the act of completing one Bond is itself the gate.

The button is always manual. There is no auto-prestige and no "are you ready" detection — pressing it is the player's commitment.

## Reset scope

Pressing Prestige executes `sim.ResetGenesis(state)`. The wipe is **deliberately heavy** — Genesis ascension is meant to feel like starting over, not like a free re-run.

| Field | Reset behaviour |
|---|---|
| `USD` | Wiped to starting value |
| `Grid` (placed components) | Wiped — grid layout cleared |
| `Owned` (component inventory) | Wiped to starting kit |
| `Research` (per-Element) | **Wiped to zero** for every Element |
| `UnlockedElements` | **Wiped** — only Hydrogen is unlocked again |
| `BinderReserves` | Wiped |
| `TokenInventory` | Wiped |
| `CurrentLoad` | Wiped |
| `Modifiers` | Rebuilt from persistent state (see below) |
| `BondsState` | **Persists** |
| `BondPoints` | **Persists** |
| `LaboratoryUpgrades` | **Persists** |
| `BestStats` | **Persists** (lifetime achievements) |
| `Layer` | Stays `LayerGenesis` for now (later layers are out of scope) |

After reset, `rebuildModifiers(state)` runs, so any Bond and Laboratory effects are immediately active for the new run. The player starts a fresh Hydrogen run with $0, no research, no unlocks — but with their compound bonuses (e.g. Methane's `+15%` $USD) already in effect.

## What stays unlocked

Bonds (synthesised compounds) and Laboratory upgrades are the only durable progression. The first prestige feels brutal because it almost is — only Methane's +15% softens Run 2. By Run 4-5, with multiple Bonds and a few Laboratory upgrades, the run feels meaningfully accelerated.

The "harder" reset (research and unlocks both wiped) is the deliberate trade-off for getting Bonds. It also makes the Laboratory upgrades that *soften* the climb — `Covalent Memory` (Helium pre-unlocked), `Stable Isotope` (research carryover), `Chain Reaction` (faster early Injectors) — into high-value early purchases. See `docs/features/0023-laboratory.md` for the full tree.

## UI surface

- **Prestige button** in the right-side panel, hidden until `len(BondsState) >= 1`.
- **Confirmation modal** on press, listing what will be reset and what will persist. Same modal pattern as the existing Hard Reset confirmation in `internal/render/settings.go`.
- **Run counter** in the panel after first prestige: "Run #N".
- **Bonds tab** (see `docs/features/0021-bonds-and-tokens.md`) becomes available the moment a player owns ≥1 Token of any element — independent of having prestiged.
- **Laboratory tab** (see `docs/features/0023-laboratory.md`) becomes visible after the first prestige.

## Save compatibility

`BondsState`, `BondPoints`, `LaboratoryUpgrades`, and `BinderReserves`/`TokenInventory` are new fields with `omitempty`. Saves from before prestige load with empty values — same gameplay as pre-prestige. No save-envelope bump per ADR 0002.

## Related

- `internal/sim/state.go` — `ResetGenesis`, persistent vs. wiped fields.
- `internal/sim/layer.go` — `Layer` enum, `LayerGenesis`.
- `internal/render/settings.go` — confirmation-modal pattern to follow.
- `docs/adr/0014-carbon-prestige-layer.md` — reset architecture, persistence model, save-compat.
- `docs/adr/0010-global-modifier-pipeline.md` — `rebuildModifiers` is reused as the post-reset modifier rebuild.
- `docs/overview.md` — Phase 4 prestige scope.
