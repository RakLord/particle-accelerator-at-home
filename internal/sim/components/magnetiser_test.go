package components

import (
	"testing"

	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/sim"
)

func TestMagnetiserAddsMagnetism(t *testing.T) {
	m := &Magnetiser{Bonus: bignum.MustParse("1.5")}
	out, lost := m.Apply(sim.ApplyContext{}, sim.Subject{Speed: 2, Magnetism: bignum.MustParse("0.5")})
	if lost {
		t.Fatal("magnetiser should never destroy subjects")
	}
	got := out.Magnetism
	if !got.Eq(bignum.FromInt(2)) {
		t.Fatalf("got %v want 2.0", got)
	}
}

func TestMagnetiserBandGate(t *testing.T) {
	m := &Magnetiser{Bonus: bignum.One()}
	// Speed=0 is below the band — inert.
	out, lost := m.Apply(sim.ApplyContext{}, sim.Subject{Speed: 0, Magnetism: bignum.Zero()})
	if lost {
		t.Fatal("magnetiser should never destroy subjects")
	}
	got := out.Magnetism
	if !got.IsZero() {
		t.Fatalf("band gate failed: got %v want 0", got)
	}
}
