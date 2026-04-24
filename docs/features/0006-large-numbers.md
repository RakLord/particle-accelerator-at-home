# Large Numbers

**Status:** live.

The economy no longer relies on native JSON numbers / `float64` for growth-oriented values. Instead it uses the large-number core in `internal/bignum`.

## Scope

Current bignum-backed fields:

- `GameState.USD`
- `Subject.Mass`
- `Subject.Magnetism`
- Element multipliers in `sim.ElementCatalog`
- Element unlock costs in `sim.ElementCatalog`
- `components.Magnetiser.Bonus`

Discrete simulation fields remain integer-backed: `Speed` is fixed-point hundredths, while `Load`, research counts, coordinates, `TickRate`, and tick counters remain plain integers.

## Arithmetic model

The bignum core uses normalized scientific decimal form.

- Non-zero values are stored as `sign × mantissa × 10^exponent`.
- `mantissa` is normalized into `[1, 10)`.
- Zero is a dedicated zero value.

Supported gameplay operations:

- add / subtract
- multiply / divide
- comparisons (`Cmp`, `LT`, `LTE`, `GT`, `GTE`, `Eq`)
- sign and zero checks

This is intentionally an incremental-game number type, not arbitrary-precision math. When two values differ by a large enough exponent gap, adding the smaller one to the larger one may leave the larger value unchanged.

## Save format

Each bignum field serializes as a canonical scientific string.

Examples:

- `"0"`
- `"1e0"`
- `"2.5e3"`
- `"7.125e-2"`

The save envelope is now version `2`. Version `1` saves are rejected rather than migrated.

## UI formatting

Display formatting is deliberately separate from save formatting.

### Default mode: scientific

- Small values render as plain decimal for readability.
- Large values render as normalized scientific notation.
- Examples: `$950`, `$12,500`, `$1.23e9`

### Alternate mode: short suffixes

- The formatter seam also supports abbreviated suffix output.
- Examples: `$950`, `$12.5K`, `$1.23B`

The current UI uses the default scientific mode, but the formatter API is structured so a player-facing setting can swap the mode later without changing save data or simulation logic.

## Comparison rules in code

Because Go cannot overload comparison operators, gameplay code must use comparison methods instead of `<`, `>`, `<=`, or `>=`.

Examples:

- `USD.GTE(cost)` for affordability checks
- `value.GT(best)` when tracking records
- `gain.IsZero()` for zero checks

## Related

- `internal/bignum/`
- `internal/sim/economy.go`
- `internal/render/header.go`
- `internal/render/periodic_table.go`
