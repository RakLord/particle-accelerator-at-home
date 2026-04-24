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

func TestHandleCodexClickSelectsInjectionElement(t *testing.T) {
	g := newTestGame()
	g.state.UnlockedElements[sim.ElementHelium] = true
	g.state.InjectionElement = sim.ElementHydrogen
	g.codexPinned = sim.ElementHelium

	bx, by, bw, bh := codexUnlockButtonRect()
	g.handleCodexClick(bx+bw/2, by+bh/2)

	if got := g.state.InjectionElement; got != sim.ElementHelium {
		t.Fatalf("InjectionElement = %q, want %q", got, sim.ElementHelium)
	}
	if g.ui.CodexNotice != "Injecting Helium" {
		t.Fatalf("CodexNotice = %q", g.ui.CodexNotice)
	}
}

func TestHandleCodexClickOutsideCardClearsPinnedElement(t *testing.T) {
	g := newTestGame()
	g.codexPinned = sim.ElementHydrogen
	g.codexHovered = sim.ElementHydrogen

	g.handleCodexClick(codexPanelX()+20, codexPanelY()+20)

	if g.codexPinned != "" {
		t.Fatalf("codexPinned = %q, want cleared", g.codexPinned)
	}
	if g.codexHovered != "" {
		t.Fatalf("codexHovered = %q, want cleared", g.codexHovered)
	}
}

func TestHandleCodexClickInsideCardKeepsPinnedElement(t *testing.T) {
	g := newTestGame()
	g.codexPinned = sim.ElementHydrogen

	g.handleCodexClick(codexCardX()+20, codexCardY()+20)

	if g.codexPinned != sim.ElementHydrogen {
		t.Fatalf("codexPinned = %q, want %q", g.codexPinned, sim.ElementHydrogen)
	}
}
