package render

import (
	"testing"

	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/ui"
)

func decimalApproxEq(a, b bignum.Decimal) bool {
	return a.Sub(b).Abs().LT(bignum.MustParse("1e-9"))
}

func TestIncomeRateWindowPerSecond(t *testing.T) {
	w := newIncomeRateWindow(sim.DefaultTickRate)
	for range incomeAverageSampleCount(sim.DefaultTickRate) {
		w.record(bignum.FromInt(2))
	}

	want := bignum.FromInt(2 * sim.DefaultTickRate)
	if got := w.perSecond(); !decimalApproxEq(got, want) {
		t.Fatalf("perSecond: got %v want %v", got, want)
	}
}

func TestIncomeRateWindowSlides(t *testing.T) {
	w := newIncomeRateWindow(sim.DefaultTickRate)
	for range incomeAverageSampleCount(sim.DefaultTickRate) {
		w.record(bignum.One())
	}
	for range sim.DefaultTickRate * 5 {
		w.record(bignum.Zero())
	}

	want := bignum.FromInt(sim.DefaultTickRate / 2)
	if got := w.perSecond(); !decimalApproxEq(got, want) {
		t.Fatalf("sliding average: got %v want %v", got, want)
	}
}

func TestHandleSettingsClickResetClearsIncomeRate(t *testing.T) {
	g := New(sim.NewGameState(), ui.NewUIState(), nil, nil)
	for range incomeAverageSampleCount(sim.DefaultTickRate) {
		g.income.record(bignum.One())
	}
	if g.income.perSecond().IsZero() {
		t.Fatalf("expected non-zero income rate before reset")
	}

	g.ui.ResetArmed = true
	g.handleSettingsClick(resetBtnX()+1, resetBtnY()+1)

	if got := g.income.perSecond(); !got.IsZero() {
		t.Fatalf("income rate should clear on reset, got %v", got)
	}
}
