package render

import (
	"testing"

	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/sim/components"
	"particleaccelerator/internal/ui"
)

// cardCenter returns the centre of the i-th inventory card.
func cardCenter(i int) (int, int) {
	x, y, w, h := invCardRect(i)
	return x + w/2, y + h/2
}

func TestInventoryCardRectsWithinModal(t *testing.T) {
	mx, my, mw, mh := invModalX(), invModalY(), invModalW, invModalH
	for i := 0; i < invCardCols*invCardRows; i++ {
		x, y, w, h := invCardRect(i)
		if x < mx || y < my || x+w > mx+mw || y+h > my+mh {
			t.Fatalf("card %d rect (%d,%d,%d,%d) escapes modal (%d,%d,%d,%d)",
				i, x, y, w, h, mx, my, mw, mh)
		}
	}
}

func TestInvToolAtMapsCardsToTools(t *testing.T) {
	for i, want := range ui.PlaceableTools {
		if i >= invCardCols*invCardRows {
			break
		}
		cx, cy := cardCenter(i)
		got := invToolAt(cx, cy)
		if got != want {
			t.Fatalf("card %d centre: got tool %v, want %v", i, got, want)
		}
	}
}

func TestInvToolAtMissesOutsideGrid(t *testing.T) {
	if got := invToolAt(0, 0); got != ui.ToolNone {
		t.Fatalf("top-left screen corner should miss, got %v", got)
	}
	// A point inside the modal header (above the card grid).
	mx := invModalX() + invModalW/2
	my := invModalY() + 20
	if got := invToolAt(mx, my); got != ui.ToolNone {
		t.Fatalf("header area should miss, got %v", got)
	}
}

func TestHandleInventoryClickSelectsAndCloses(t *testing.T) {
	g := newTestGame()
	g.ui.InventoryOpen = true
	// Click card #0 (Injector) — unlocked, should select + close.
	cx, cy := cardCenter(0)
	g.handleInventoryClick(cx, cy)
	if g.ui.InventoryOpen {
		t.Fatal("clicking unlocked card should close the modal")
	}
	if g.ui.Selected != ui.ToolInjector {
		t.Fatalf("expected ToolInjector selected, got %v", g.ui.Selected)
	}
	if g.ui.InventoryHovered != ui.ToolNone {
		t.Fatalf("hovered should reset on close, got %v", g.ui.InventoryHovered)
	}
}

func TestHandleInventoryClickCloseButton(t *testing.T) {
	g := newTestGame()
	g.ui.InventoryOpen = true
	g.ui.Selected = ui.ToolAccelerator // arbitrary prior selection
	cx := invCloseX() + closeBtnW/2
	cy := invCloseY() + closeBtnH/2
	g.handleInventoryClick(cx, cy)
	if g.ui.InventoryOpen {
		t.Fatal("close button should dismiss the modal")
	}
	if g.ui.Selected != ui.ToolAccelerator {
		t.Fatalf("close button should not change selection, got %v", g.ui.Selected)
	}
}

func TestHandleInventoryClickOutsidePanelCloses(t *testing.T) {
	g := newTestGame()
	g.ui.InventoryOpen = true
	g.handleInventoryClick(1, 1) // top-left corner, well outside panel
	if g.ui.InventoryOpen {
		t.Fatal("clicking outside the modal panel should dismiss it")
	}
}

func TestHandleLogClickCloseButton(t *testing.T) {
	g := newTestGame()
	g.ui.LogOpen = true
	g.handleLogClick(logCloseX()+closeBtnW/2, logCloseY()+closeBtnH/2)
	if g.ui.LogOpen {
		t.Fatal("close button should dismiss log modal")
	}
}

func TestHandleLogClickOutsidePanelCloses(t *testing.T) {
	g := newTestGame()
	g.ui.LogOpen = true
	g.handleLogClick(1, 1)
	if g.ui.LogOpen {
		t.Fatal("clicking outside log modal should dismiss it")
	}
}

func TestHandleLogClickInsidePanelKeepsOpen(t *testing.T) {
	g := newTestGame()
	g.ui.LogOpen = true
	g.handleLogClick(logModalX()+20, logModalY()+80)
	if !g.ui.LogOpen {
		t.Fatal("clicking inside log modal body should keep it open")
	}
}

func TestLoadBarFillWidthClamps(t *testing.T) {
	if got := loadBarFillWidth(0, 16, 100); got != 0 {
		t.Fatalf("empty load fill = %d, want 0", got)
	}
	if got := loadBarFillWidth(8, 16, 100); got != 50 {
		t.Fatalf("half load fill = %d, want 50", got)
	}
	if got := loadBarFillWidth(20, 16, 100); got != 100 {
		t.Fatalf("over cap load fill = %d, want 100", got)
	}
	if got := loadBarFillWidth(1, 0, 100); got != 0 {
		t.Fatalf("zero cap load fill = %d, want 0", got)
	}
}

func TestInjectButtonLabelStates(t *testing.T) {
	s := sim.NewGameState()
	label, enabled := injectButtonLabel(s)
	if enabled || label != "No Injector" {
		t.Fatalf("no injector label=%q enabled=%v", label, enabled)
	}

	s.Grid.Cells[0][0].Component = &components.Injector{Direction: sim.DirEast}
	label, enabled = injectButtonLabel(s)
	if !enabled || label != "Inject" {
		t.Fatalf("ready label=%q enabled=%v", label, enabled)
	}

	s.InjectionCooldownRemaining = s.TickRate * 2
	label, enabled = injectButtonLabel(s)
	if enabled || label != "Cooldown 2s" {
		t.Fatalf("cooldown label=%q enabled=%v", label, enabled)
	}

	s.InjectionCooldownRemaining = 0
	s.CurrentLoad = s.EffectiveMaxLoad()
	label, enabled = injectButtonLabel(s)
	if enabled || label != "Max Load" {
		t.Fatalf("max load label=%q enabled=%v", label, enabled)
	}
}

// newTestGame builds a Game wired to a fresh sim state, no save hooks.
func newTestGame() *Game {
	return New(sim.NewGameState(), ui.NewUIState(), nil, nil)
}
