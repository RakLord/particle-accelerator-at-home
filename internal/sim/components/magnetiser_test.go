package components

import (
	"testing"

	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/sim"
)

func TestMagnetiserAddsMagnetism(t *testing.T) {
	m := &Magnetiser{Bonus: bignum.MustParse("1.5")}
	got := m.Apply(sim.Subject{Speed: 2, Magnetism: bignum.MustParse("0.5")}).Magnetism
	if !got.Eq(bignum.FromInt(2)) {
		t.Fatalf("got %v want 2.0", got)
	}
}

func TestMagnetiserBandGate(t *testing.T) {
	m := &Magnetiser{Bonus: bignum.One()}
	// Speed=0 is below the band — inert.
	got := m.Apply(sim.Subject{Speed: 0, Magnetism: bignum.Zero()}).Magnetism
	if !got.IsZero() {
		t.Fatalf("band gate failed: got %v want 0", got)
	}
}
