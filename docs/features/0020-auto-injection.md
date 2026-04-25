# Auto-Injection

**Status:** Phase 4.

## Concept

Auto-Injection is an idle-mode injection toggle. While on, the manual Inject path fires automatically every cooldown cycle without the player pressing the button. It is the prestige layer's idle-game enabler — the player can step away from the keyboard and Subjects keep flowing.

The mechanic is **gated behind the Benzene Bond** (`docs/features/0021-bonds-and-tokens.md`). Until the player has synthesised Benzene, the Auto-Inject toggle is hidden. Synthesising Benzene reveals the toggle for all current and future runs.

## Behaviour

When `GameState.Modifiers.AutoInjectEnabled == true` **and** the player has the Auto-Inject toggle set to ON, the simulation invokes `state.Inject()` automatically every `AutoInjectCadence` ticks.

The auto-fire path reuses the same admission logic as manual Inject:

- Respects global injection cooldown.
- Respects Max Load (via `EffectiveMaxLoad()`).
- Skips if no Injectors are placed.

There is no separate auto-inject queue — the auto path is a tick-driven caller of `Inject()`, identical to the player pressing the button. `InjectorRateMul` (the `Acetylene` Bond's effect) still applies.

### Cadence

Default Auto-Inject cadence on Benzene unlock is **10 seconds** — slower than the 5-second base manual cooldown. An active player out-paces auto by a factor of 2; auto is a "set it and walk away" tool, not the optimal play.

The Laboratory tree (see `docs/features/0023-laboratory.md`) sells four cadence reductions:

| Lab upgrade | Effect | Bond Point cost |
|---|---|---:|
| Auto-Inject Speed I | 10s → 8s | 1 |
| Auto-Inject Speed II | 8s → 6s | 2 |
| Auto-Inject Speed III | 6s → 5s (matches manual base) | 3 |
| Auto-Inject Speed IV | 5s → 4s (faster than manual base) | 4 |

Speed IV is intentionally cheaper than buying it via Acetylene + manual play would imply: the player has invested heavily in the prestige loop to reach this depth, and the reward is a true idle-friendly cadence. With both Acetylene and Speed IV, manual play remains slightly faster (4s ÷ 1.333 ≈ 3s effective), so the active vs. idle tension stays alive.

The Lab levels stack; only the highest purchased level applies. Level is read from `LaboratoryUpgrades["auto_inject_speed"]` (an int, capped at 4).

### Toggle UI

When Benzene is synthesised, an `Auto Inject [ON|OFF]` toggle appears in the right-side panel next to the existing Inject button. Default OFF on first Benzene synthesis; persistence across runs is handled by `GameState.AutoInjectActive bool` (a UI-only preference, not a modifier).

The toggle is not gated by run state — once Benzene is owned, it's available immediately on every Run including post-prestige Run 1. (Even though the player has no Injectors placed yet on a fresh prestige run, the toggle costs nothing to leave on; it's a no-op until Injectors exist.)

### Composition with Acetylene and Chain Reaction

Auto-Inject cadence is **independent** of `InjectorRateMul` (Acetylene's manual cooldown speed-up) and `Chain Reaction` (the Lab upgrade that doubles Injector base emission rate). Those modifiers shape what each Inject *call* does; Auto-Inject only schedules the calls.

Concretely:

- Without Acetylene, Auto-Inject Speed III: every 5s, manual Inject fires.
- With Acetylene + Auto-Inject Speed III: every 5s, manual Inject fires *and* the manual cooldown was already 25% shorter — so the next manual press is also faster. The auto path doesn't get faster from Acetylene because its cadence is the *scheduling* interval, not the cooldown.

This separation is intentional. If the two modifiers stacked, Speed IV + Acetylene would push the auto cadence below 1 tick and break the simulation.

## Implementation surface

A new field on `GlobalModifiers`:

```go
AutoInjectEnabled bool   // set by Benzene Bond's Apply closure
```

A new int on `GlobalModifiers`:

```go
AutoInjectCadenceTicks int  // computed from LaboratoryUpgrades level + base
```

The tick loop, after the existing manual-inject cooldown advance, runs:

```go
if state.Modifiers.AutoInjectEnabled && state.AutoInjectActive {
    state.AutoInjectTickCounter++
    if state.AutoInjectTickCounter >= state.Modifiers.AutoInjectCadenceTicks {
        state.AutoInjectTickCounter = 0
        _ = state.Inject() // ignore admission errors; auto path is best-effort
    }
}
```

`AutoInjectTickCounter` is on `GameState` (so it persists across saves and pauses correctly) but is **not** in the save schema — it's transient. `AutoInjectActive` is in the save schema so the toggle state persists.

## Save compatibility

`AutoInjectActive bool` is a new field with `omitempty`. Saves without it default to `false` (toggle off), which is also the post-Benzene first-time default. No save-envelope bump.

## Related

- `internal/sim/injection.go` — auto-inject tick loop.
- `internal/sim/bonds.go` — Benzene's `Apply` closure sets `AutoInjectEnabled`.
- `internal/sim/laboratory.go` — Speed I-IV upgrades drive `AutoInjectCadenceTicks`.
- `internal/sim/modifiers.go` — `AutoInjectEnabled`, `AutoInjectCadenceTicks` fields.
- `internal/ui/inject_panel.go` — toggle UI (extends the existing Inject button area).
- `docs/adr/0017-auto-injection.md` — separation from `InjectorRateMul`, scheduling model, modifier fields.
- `docs/features/0015-manual-injection.md` — the path auto-inject reuses.
- `docs/features/0021-bonds-and-tokens.md` — Benzene Bond is the unlock.
- `docs/features/0023-laboratory.md` — cadence upgrades.
