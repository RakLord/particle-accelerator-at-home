package render

import (
	"strings"
	"testing"

	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/ui"
)

func TestFirstFiveUSDMilestoneShowsAndLogsOnce(t *testing.T) {
	g := newHelperTestGame()
	g.state.USD = bignum.FromInt(5)

	g.checkHelperMilestones()
	if !g.ui.HelperOpen {
		t.Fatal("first $5 milestone should open helper")
	}
	if !g.state.HasShownHelperMilestone(milestoneFirstFiveUSD) {
		t.Fatal("first $5 milestone should be marked shown")
	}
	if got := len(g.state.NotificationLog); got != 1 {
		t.Fatalf("NotificationLog length = %d, want 1", got)
	}

	g.closeHelper()
	g.checkHelperMilestones()
	if g.ui.HelperOpen {
		t.Fatal("first $5 milestone should not reopen after being shown")
	}
	if got := len(g.state.NotificationLog); got != 1 {
		t.Fatalf("NotificationLog length after second check = %d, want 1", got)
	}
}

func TestComponentHelpDoesNotLog(t *testing.T) {
	g := newHelperTestGame()
	pos := sim.Position{X: 2, Y: 2}
	g.state.Grid.Cells[pos.Y][pos.X].IsCollector = true

	if !g.showComponentHelpAt(pos, 640, 360) {
		t.Fatal("collector help should open")
	}
	if !g.ui.HelperOpen {
		t.Fatal("component help should set HelperOpen")
	}
	if got := len(g.state.NotificationLog); got != 0 {
		t.Fatalf("component help should not log, got %d entries", got)
	}
	if !strings.Contains(g.ui.HelperBody, "Next cost:") {
		t.Fatalf("component help should include live stats, got %q", g.ui.HelperBody)
	}
}

func TestCursorHelperRectStaysOnScreen(t *testing.T) {
	g := newHelperTestGame()
	g.ui.HelperOpen = true
	g.ui.HelperHeader = "Help"
	g.ui.HelperBody = strings.Repeat("long body ", 80)
	g.ui.HelperX = screenW - 2
	g.ui.HelperY = screenH - 2

	x, y, w, h := helperModalRect(g.ui)
	if x < 0 || y < 0 || x+w > screenW || y+h > screenH {
		t.Fatalf("helper rect escaped screen: (%d,%d,%d,%d)", x, y, w, h)
	}
}

func newHelperTestGame() *Game {
	return New(sim.NewGameState(), ui.NewUIState(), nil, nil)
}
