# ADR 0004 — Large-number core and canonical save strings

**Status:** accepted.
**Date:** 2026-04-22.

## Context

Phase 2 still stored `$USD`, Subject `Mass`, Subject `Magnetism`, Element value multipliers, and Element unlock costs as `float64`. That was acceptable for the first playable loop, but it creates three problems for an incremental game:

1. `float64` tops out around `1e308`, well below the range an idle economy will eventually reach.
2. Save JSON encoded these values as native numbers, coupling persistence to Go's current in-memory representation and making future formatter changes harder.
3. The game ships on both desktop and WASM. A browser-only dependency such as `break_infinity.js` would not solve the desktop path and would force the sim through a JS interop seam.

The replacement needs to stay fast enough for per-tick economy math, be easy to compare in game logic (`USD >= cost`), and serialize cleanly.

## Decision

**1. Add a dedicated cross-platform large-number package: `internal/bignum`.**
- The core type stores a normalized scientific-decimal representation: sign, mantissa, exponent.
- Zero is a dedicated zero value.
- Normalized invariant: non-zero values keep `mantissa` in `[1, 10)` and an integer base-10 exponent.
- The package optimizes for incremental-game workloads: fast compare, multiply, divide, and "good enough" addition/subtraction, not arbitrary-precision exact math.

**2. Use method-based comparison, not operators.**
- Go cannot overload `<`, `>`, `<=`, or `>=`.
- The package exposes `Cmp`, `Eq`, `LT`, `LTE`, `GT`, `GTE`, `IsZero`, and `Sign`.
- Game logic uses these methods anywhere it previously relied on native numeric operators.

**3. Migrate growth-oriented gameplay scalars to `bignum`.**
- `GameState.USD`
- `Subject.Mass`
- `Subject.Magnetism`
- `ElementInfo.Multiplier`
- `ElementInfo.UnlockCost`
- `components.Magnetiser.Bonus`

Structural integers remain native integers: `Speed`, `Load`, research counts, grid coordinates, `TickRate`, and tick counters. They drive discrete simulation flow rather than large-value progression.

**4. Persist bignums as canonical scientific strings.**
- JSON encodes a bignum as a string such as `"2.5e3"` or `"1e0"`.
- Save envelope version bumps from `1` to `2`.
- Version `1` saves are intentionally rejected rather than migrated.

**5. Keep display formatting separate from save formatting.**
- `String()` / JSON serialization remain canonical and stable for save/load.
- Player-facing display goes through formatter helpers.
- Default display mode is scientific for large magnitudes, with plain decimal output for comfortably small values.
- A suffix formatter is part of the same seam so a future player setting can switch display style without touching save data or sim math.

## Consequences

**Wins**
- Economy values are no longer capped by `float64`'s exponent range.
- Save data becomes representation-stable and easy to parse across Go and JavaScript.
- Comparison-heavy game logic stays explicit and cheap.
- UI display policy is now a layer that can change independently from persistence.

**Costs**
- Addition/subtraction lose precision when magnitudes differ heavily. Acceptable for idle-game economy math and consistent with the performance goal.
- Existing version `1` saves no longer load.
- More call sites must use methods (`Add`, `Cmp`, `LT`, etc.) instead of native operators.

## Alternatives considered

- **Keep `float64` and switch only the renderer to scientific notation.** Rejected: formatting alone does not solve overflow or precision loss.
- **Call `break_infinity.js` from Go/WASM.** Rejected: the desktop build would still need a second implementation, and the sim would gain a JS boundary in a hot path.
- **Use `math/big`.** Rejected: unnecessary accuracy and allocation overhead for an incremental game's common-case math.

## Related

- `internal/bignum/` — large-number representation, arithmetic, comparison, formatting, JSON encoding.
- `internal/sim/economy.go` — value formula and costs now operate on bignums.
- `internal/sim/save.go` — save envelope bumped to v2.
- `docs/features/large-numbers.md` — gameplay-facing behavior and UI format rules.
- ADR-0002 — versioned save schema; this change is a breaking save-shape update.
