# Manual Injection and Load Bar

**Status:** Phase 3.

Injection is player-triggered. Injector components are placed spawn sources, but they do not emit automatically by default.

## Inject action

The right-side panel contains an `Inject` button. Pressing it commands every placed Injector to emit one Subject in grid scan order.

The action succeeds only when:

- at least one Injector is placed
- the global injection cooldown is ready
- current Load is below effective Max Load

If at least one Subject is admitted, the global cooldown starts. The base cooldown is 5 seconds, converted to ticks from `GameState.TickRate`. `GlobalModifiers.InjectorRateMul` shortens the effective cooldown so future injection-speed upgrades can reuse the same path. Auto-injection is intentionally not enabled yet, but can later call the same `GameState.Inject` path.

## Load bar

The grid view renders `Load: Current/Max` above the accelerator as a progress bar. It uses `GameState.EffectiveMaxLoad()` so future Max Load bonuses display correctly.

Max Load admission is enforced per emitted Subject. If several Injectors fire but only some fit, the ones that fit are admitted and the rest are skipped.
