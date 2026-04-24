package render

import (
	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/sim"
)

const numberDisplayMode = bignum.DisplayScientific

func formatUSD(v bignum.Decimal) string {
	return "$" + v.Format(numberDisplayMode, 2)
}

func formatIncomeRate(v bignum.Decimal) string {
	if v.Sign() < 0 {
		return "-" + formatUSD(v.Abs()) + "/s"
	}
	return "+" + formatUSD(v) + "/s"
}

func formatMultiplier(v bignum.Decimal) string {
	return "×" + v.Format(numberDisplayMode, 2)
}

func formatNumber(v bignum.Decimal) string {
	return v.Format(numberDisplayMode, 2)
}

func formatSpeed(v sim.Speed) string {
	return v.Decimal().Format(numberDisplayMode, 2)
}
