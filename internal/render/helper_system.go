package render

import (
	"strings"
	"time"

	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/ui"
)

const milestoneFirstFiveUSD = "first-five-usd"

var milestoneFirstFiveUSDThreshold = bignum.MustParse("5")

func (g *Game) checkHelperMilestones() {
	if g.state.HasShownHelperMilestone(milestoneFirstFiveUSD) || g.state.USD.LT(milestoneFirstFiveUSDThreshold) {
		return
	}
	g.showHelper(
		"Inventory Available",
		"You can press E to open the Inventory. In the Inventory you can buy new components for your Accelerator.",
		true,
		0,
		0,
		true,
		milestoneFirstFiveUSD,
	)
}

func (g *Game) showHelper(header, body string, centered bool, x, y int, logNotification bool, milestoneID string) bool {
	if g.ui.HelperOpen {
		return false
	}
	if milestoneID != "" && g.state.HasShownHelperMilestone(milestoneID) {
		return false
	}
	g.ui.HelperOpen = true
	g.ui.HelperHeader = header
	g.ui.HelperBody = body
	g.ui.HelperCentered = centered
	g.ui.HelperX = x
	g.ui.HelperY = y
	if logNotification {
		g.state.RecordNotification(header, body, time.Now().Format("15:04"))
	}
	if milestoneID != "" {
		g.state.MarkHelperMilestoneShown(milestoneID)
	}
	return true
}

func (g *Game) showComponentHelpAt(pos sim.Position, mx, my int) bool {
	cell := g.state.Grid.Cells[pos.Y][pos.X]
	tool, kind, ok := toolForCell(cell)
	if !ok {
		return false
	}
	info, ok := ui.ToolInfoCatalog[tool]
	if !ok {
		return false
	}

	parts := []string{}
	if info.Tagline != "" {
		parts = append(parts, info.Tagline)
	}
	if info.Description != "" {
		parts = append(parts, info.Description)
	}
	if kind != "" {
		parts = append(parts, g.componentHelpStats(kind))
	}
	return g.showHelper(info.Name, strings.Join(parts, "\n\n"), false, mx, my, false, "")
}

func (g *Game) componentHelpStats(kind sim.ComponentKind) string {
	tier := sim.BaseTier
	if g.state.ComponentTiers != nil {
		if t, ok := g.state.ComponentTiers[kind]; ok && t >= sim.BaseTier {
			tier = t
		}
	}
	owned := 0
	if g.state.Owned != nil {
		owned = g.state.Owned[kind]
	}
	return "Tier: " + itoa(int(tier)) +
		"\nOwned: " + itoa(owned) +
		"\nPlaced: " + itoa(sim.CountPlaced(g.state, kind)) +
		"\nAvailable: " + itoa(sim.CountAvailable(g.state, kind)) +
		"\nNext cost: " + formatUSD(sim.ComponentCost(g.state, kind))
}

func toolForCell(cell sim.Cell) (ui.Tool, sim.ComponentKind, bool) {
	if cell.IsCollector {
		return ui.ToolCollector, sim.KindCollector, true
	}
	if cell.Component == nil {
		return ui.ToolNone, "", false
	}
	kind := cell.Component.Kind()
	switch kind {
	case sim.KindInjector:
		return ui.ToolInjector, kind, true
	case sim.KindAccelerator:
		return ui.ToolAccelerator, kind, true
	case sim.KindMeshGrid:
		return ui.ToolMeshGrid, kind, true
	case sim.KindMagnetiser:
		return ui.ToolMagnetiser, kind, true
	case sim.KindRotator:
		return ui.ToolElbow, kind, true
	case sim.KindResonator:
		return ui.ToolResonator, kind, true
	case sim.KindCatalyst:
		return ui.ToolCatalyst, kind, true
	case sim.KindDuplicator:
		return ui.ToolDuplicator, kind, true
	}
	return ui.ToolNone, "", false
}
