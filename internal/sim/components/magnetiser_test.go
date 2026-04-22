package components

import (
	"testing"

	"particleaccelerator/internal/sim"
)

func TestMagnetiserAddsMagnetism(t *testing.T) {
	m := &Magnetiser{Bonus: 1.5}
	got := m.Apply(sim.Subject{Speed: 2, Magnetism: 0.5}).Magnetism
	if got != 2.0 {
		t.Fatalf("got %v want 2.0", got)
	}
}

func TestMagnetiserBandGate(t *testing.T) {
	m := &Magnetiser{Bonus: 1}
	// Speed=0 is below the band — inert.
	got := m.Apply(sim.Subject{Speed: 0, Magnetism: 0}).Magnetism
	if got != 0 {
		t.Fatalf("band gate failed: got %v want 0", got)
	}
}
