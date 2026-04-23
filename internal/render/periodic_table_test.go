package render

import (
	"testing"

	"particleaccelerator/internal/sim"
)

func TestCodexTileRectUsesPeriodicCoordinates(t *testing.T) {
	hx, hy, hw, hh := codexTileRect(sim.ElementHydrogen)
	if hw != codexTileSize || hh != codexTileSize {
		t.Fatalf("Hydrogen tile size: got %dx%d want %dx%d", hw, hh, codexTileSize, codexTileSize)
	}
	hex, hey, _, _ := codexTileRect(sim.ElementHelium)
	if hy != hey {
		t.Fatalf("expected Hydrogen and Helium on same row: hy=%d hey=%d", hy, hey)
	}
	if hx >= hex {
		t.Fatalf("expected Hydrogen left of Helium: hx=%d hex=%d", hx, hex)
	}
}

func TestCodexElementAtFindsTileCenters(t *testing.T) {
	hx, hy, hw, hh := codexTileRect(sim.ElementHydrogen)
	if got, ok := codexElementAt(hx+hw/2, hy+hh/2); !ok || got != sim.ElementHydrogen {
		t.Fatalf("expected Hydrogen hit, got %q ok=%v", got, ok)
	}
	hex, hey, hew, heh := codexTileRect(sim.ElementHelium)
	if got, ok := codexElementAt(hex+hew/2, hey+heh/2); !ok || got != sim.ElementHelium {
		t.Fatalf("expected Helium hit, got %q ok=%v", got, ok)
	}
}

func TestCodexFocusedElementPinnedOverridesHover(t *testing.T) {
	if got := codexFocusedElement(sim.ElementHydrogen, sim.ElementHelium); got != sim.ElementHelium {
		t.Fatalf("expected pinned element to win, got %q", got)
	}
	if got := codexFocusedElement(sim.ElementHydrogen, ""); got != sim.ElementHydrogen {
		t.Fatalf("expected hover element when nothing pinned, got %q", got)
	}
}
