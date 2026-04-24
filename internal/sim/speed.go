package sim

import "particleaccelerator/internal/bignum"

// Speed is fixed-point hundredths of the player-facing Speed value.
// SpeedFromInt(1) is displayed as 1.00 and moves one cell every SpeedDivisor
// ticks. Keeping this as an integer preserves deterministic movement while
// allowing fractional speeds such as 0.50 after Mesh Grid throttling.
type Speed int64

const (
	SpeedScale Speed = 100
	MinSpeed   Speed = 1
)

// StepProgressPerCell is the fixed-point movement accumulated to cross one
// grid cell.
const StepProgressPerCell Speed = SpeedScale * SpeedDivisor

func SpeedFromInt(v int) Speed { return Speed(v) * SpeedScale }

func SpeedFromRatio(numerator, denominator int) Speed {
	if denominator == 0 {
		panic("sim: speed ratio denominator is zero")
	}
	return Speed(numerator) * SpeedScale / Speed(denominator)
}

func (s Speed) Decimal() bignum.Decimal {
	return bignum.FromInt64(int64(s)).DivInt(int(SpeedScale))
}

func (s Speed) ScaledFromLegacy() Speed { return s * SpeedScale }

func ClampMinSpeed(s Speed) Speed {
	if s < MinSpeed {
		return MinSpeed
	}
	return s
}
