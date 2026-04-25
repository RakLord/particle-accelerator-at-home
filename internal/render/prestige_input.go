package render

import (
	"log"

	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/ui"
)

func (g *Game) handlePrestigePanelClick(mx, my int) bool {
	px, py, pw, ph := prestigeButtonRect()
	if contains(mx, my, px, py, pw, ph) && sim.HasAnyBond(g.state) {
		g.ui.PrestigeConfirmOpen = true
		g.ui.PrestigeNotice = ""
		return true
	}
	tsx, tsy, tsw, tsh := storeTabRect()
	if contains(mx, my, tsx, tsy, tsw, tsh) {
		g.ui.PrestigeTab = ui.PrestigeTabStore
		g.ui.PrestigeNotice = ""
		return true
	}
	bx, by, bw, bh := bondsTabRect()
	if canShowBondsTab(g.state) && contains(mx, my, bx, by, bw, bh) {
		g.ui.PrestigeTab = ui.PrestigeTabBonds
		g.ui.PrestigeNotice = ""
		return true
	}
	if !contains(mx, my, prestigePanelX, prestigePanelY, prestigePanelW, prestigePanelH) {
		return false
	}
	if g.ui.PrestigeTab == ui.PrestigeTabBonds && canShowBondsTab(g.state) {
		return g.handleBondClick(mx, my)
	}
	return g.handleBinderStoreClick(mx, my)
}

func (g *Game) handleBinderStoreClick(mx, my int) bool {
	for i, e := range sim.BinderStoreElementOrder {
		bx, by, bw, bh := crystalliseButtonRect(i)
		if !contains(mx, my, bx, by, bw, bh) {
			continue
		}
		if err := sim.CrystalliseToken(g.state, e); err != nil {
			g.ui.PrestigeNotice = "Need more " + sim.ElementCatalog[e].Symbol + " reserve"
			return true
		}
		g.ui.PrestigeNotice = sim.ElementCatalog[e].Symbol + " Token crystallised"
		return true
	}
	return true
}

func (g *Game) handleBondClick(mx, my int) bool {
	for i, id := range sim.BondCatalogOrder {
		bx, by, bw, bh := synthesiseButtonRect(i)
		if !contains(mx, my, bx, by, bw, bh) {
			continue
		}
		bond := sim.BondCatalog[id]
		if err := sim.SynthesiseBond(g.state, id); err != nil {
			g.ui.PrestigeNotice = "Need Tokens for " + bond.Name
			return true
		}
		g.ui.PrestigeNotice = bond.Name + " synthesised"
		return true
	}
	return true
}

func (g *Game) handlePrestigeConfirmClick(mx, my int) {
	cx, cy, cw, ch := prestigeCancelButtonRect()
	if contains(mx, my, cx, cy, cw, ch) || !contains(mx, my, prestigeConfirmX(), prestigeConfirmY(), prestigeConfirmW, prestigeConfirmH) {
		g.ui.PrestigeConfirmOpen = false
		return
	}
	px, py, pw, ph := prestigeConfirmButtonRect()
	if !contains(mx, my, px, py, pw, ph) {
		return
	}
	sim.ResetGenesis(g.state)
	g.income.reset(g.state.TickRate)
	g.trail = g.trail[:0]
	g.ticksSinceSave = 0
	g.ui.PrestigeConfirmOpen = false
	g.ui.PrestigeTab = ui.PrestigeTabStore
	g.ui.PrestigeNotice = "Prestiged to Run #" + itoa(g.state.RunCount+1)
	if g.save == nil {
		return
	}
	if err := g.save(); err != nil {
		log.Printf("prestige save failed: %v", err)
		g.ui.AutosaveError = err.Error()
		g.ui.PrestigeNotice = "Prestige save failed: " + err.Error()
		return
	}
	g.ui.AutosaveError = ""
}
