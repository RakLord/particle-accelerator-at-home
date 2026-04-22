package render

import (
	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/sim"
)

const incomeAverageWindowSeconds = 10

// incomeRateWindow tracks a fixed 10-second rolling sum of collector income.
// Empty slots stay zero during startup so the display stays smoothed from tick 1.
type incomeRateWindow struct {
	samples []bignum.Decimal
	next    int
	sum     bignum.Decimal
}

func newIncomeRateWindow(tickRate int) incomeRateWindow {
	w := incomeRateWindow{}
	w.reset(tickRate)
	return w
}

func (w *incomeRateWindow) reset(tickRate int) {
	w.samples = make([]bignum.Decimal, incomeAverageSampleCount(tickRate))
	w.next = 0
	w.sum = bignum.Zero()
}

func (w *incomeRateWindow) ensureTickRate(tickRate int) {
	if len(w.samples) == incomeAverageSampleCount(tickRate) {
		return
	}
	w.reset(tickRate)
}

func (w *incomeRateWindow) record(delta bignum.Decimal) {
	if len(w.samples) == 0 {
		return
	}
	w.sum = w.sum.Sub(w.samples[w.next]).Add(delta)
	w.samples[w.next] = delta
	w.next++
	if w.next == len(w.samples) {
		w.next = 0
	}
}

func (w incomeRateWindow) perSecond() bignum.Decimal {
	if len(w.samples) == 0 {
		return bignum.Zero()
	}
	return w.sum.DivInt(incomeAverageWindowSeconds)
}

func incomeAverageSampleCount(tickRate int) int {
	if tickRate <= 0 {
		tickRate = sim.DefaultTickRate
	}
	return tickRate * incomeAverageWindowSeconds
}
