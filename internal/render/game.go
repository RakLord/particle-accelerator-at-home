package render

import (
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"particleaccelerator/internal/input"
	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/ui"
)

// SaveFn is invoked when the player presses "Save now" in the settings modal.
type SaveFn func() error

// ResetFn is invoked after the player confirms a hard reset. Implementations
// should wipe persistent save data so a subsequent restart doesn't restore
// the pre-reset state.
type ResetFn func() error

const autosaveInterval = 300

type Game struct {
	state          *sim.GameState
	ui             *ui.UIState
	save           SaveFn
	reset          ResetFn
	ticksSinceSave int
	income         incomeRateWindow

	// Render interpolation. lastTickAt is set right after each sim Tick and
	// Draw uses (now - lastTickAt)/tickDuration as the interpolation alpha.
	lastTickAt   time.Time
	tickDuration time.Duration

	// Particle trail samples, rendered below live Subjects. Session-scoped
	// (not persisted) and cleared when the user toggles trails off.
	trail []trailSample

	// Codex interaction state is render-local session state. It controls which
	// element card is shown while the Codex overlay is open and is intentionally
	// not persisted.
	codexHovered sim.Element
	codexPinned  sim.Element
}

func New(s *sim.GameState, u *ui.UIState, save SaveFn, reset ResetFn) *Game {
	return &Game{
		state:        s,
		ui:           u,
		save:         save,
		reset:        reset,
		income:       newIncomeRateWindow(s.TickRate),
		lastTickAt:   time.Now(),
		tickDuration: tickDurationFor(s.TickRate),
	}
}

// tickDurationFor returns the wall-clock duration of a single sim tick at the
// given TPS. Guards against TickRate = 0 (e.g. malformed save) to avoid a
// divide-by-zero in the alpha calculation.
func tickDurationFor(tps int) time.Duration {
	if tps <= 0 {
		tps = sim.DefaultTickRate
	}
	return time.Second / time.Duration(tps)
}

func (g *Game) Update() error {
	g.handleInput()
	beforeUSD := g.state.USD
	g.state.Tick()
	g.income.ensureTickRate(g.state.TickRate)
	g.income.record(g.state.USD.Sub(beforeUSD))
	g.checkHelperMilestones()
	g.lastTickAt = time.Now()
	g.tickDuration = tickDurationFor(g.state.TickRate)
	g.ticksSinceSave++
	if g.ticksSinceSave >= autosaveInterval && g.save != nil {
		if err := g.save(); err != nil {
			log.Printf("autosave failed: %v", err)
			g.ui.AutosaveError = err.Error()
		} else {
			g.ui.AutosaveError = ""
		}
		g.ticksSinceSave = 0
	}
	return nil
}

func (g *Game) handleInput() {
	mx, my := ebiten.CursorPosition()

	// Global: Escape closes the topmost overlay before any other input is handled.
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if g.ui.PrestigeConfirmOpen {
			g.ui.PrestigeConfirmOpen = false
			return
		}
		if g.ui.HelperOpen {
			g.closeHelper()
			return
		}
		if g.ui.HotkeysOpen {
			g.ui.HotkeysOpen = false
			return
		}
		if g.ui.NotificationHistoryOpen {
			g.ui.NotificationHistoryOpen = false
			return
		}
		g.closeModals()
		return
	}
	if g.ui.HelperOpen {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && helperCloseHit(g.ui, mx, my) {
			g.closeHelper()
		}
		return
	}
	if g.ui.NotificationHistoryOpen {
		_, wheelY := ebiten.Wheel()
		if wheelY > 0 {
			g.ui.NotificationHistoryScroll--
		}
		if wheelY < 0 {
			g.ui.NotificationHistoryScroll++
		}
		if g.ui.NotificationHistoryScroll < 0 {
			g.ui.NotificationHistoryScroll = 0
		}
		if maxScroll := maxNotificationScroll(g.state); g.ui.NotificationHistoryScroll > maxScroll {
			g.ui.NotificationHistoryScroll = maxScroll
		}
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.handleNotificationHistoryClick(mx, my)
		}
		return
	}
	if g.ui.HotkeysOpen {
		if inpututil.IsKeyJustPressed(ebiten.KeySlash) {
			g.ui.HotkeysOpen = false
			return
		}
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.handleHotkeysClick(mx, my)
		}
		return
	}
	if g.ui.PrestigeConfirmOpen {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.handlePrestigeConfirmClick(mx, my)
		}
		return
	}

	// Global: toggle particle trails with T. Clears the buffer on disable so
	// old dots don't linger for their full lifetime after toggle-off.
	if inpututil.IsKeyJustPressed(ebiten.KeyT) {
		g.ui.TrailsEnabled = !g.ui.TrailsEnabled
		if !g.ui.TrailsEnabled {
			g.trail = g.trail[:0]
		}
	}

	// Global: modal hotkeys.
	if inpututil.IsKeyJustPressed(ebiten.KeyL) {
		g.toggleLog()
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		g.toggleCodex()
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySlash) {
		g.toggleHotkeys()
		return
	}

	// Global: toggle the inventory modal with E. Closes any other modal that
	// may be open so the player isn't stuck with two overlays.
	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		g.ui.InventoryOpen = !g.ui.InventoryOpen
		g.ui.InventoryHovered = ui.ToolNone
		if g.ui.InventoryOpen {
			g.ui.SettingsOpen = false
			g.ui.CodexOpen = false
			g.ui.LogOpen = false
			g.ui.NotificationHistoryOpen = false
			g.ui.HotkeysOpen = false
		}
		return
	}

	// Modals swallow clicks when open.
	if g.ui.SettingsOpen {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.handleSettingsClick(mx, my)
		}
		return
	}
	if g.ui.CodexOpen {
		g.updateCodexHover(mx, my)
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.handleCodexClick(mx, my)
		}
		return
	}
	if g.ui.InventoryOpen {
		g.ui.InventoryHovered = invToolAt(mx, my)
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.handleInventoryClick(mx, my)
		}
		return
	}
	if g.ui.LogOpen {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.handleLogClick(mx, my)
		}
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.triggerInjection()
		return
	}

	// Header: inventory + log + codex + settings buttons.
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if contains(mx, my, settingsBtnX, settingsBtnY, settingsBtnW, settingsBtnH) {
			g.openSettings()
			return
		}
		if contains(mx, my, codexBtnX, codexBtnY, codexBtnW, codexBtnH) {
			g.openCodex()
			return
		}
		if contains(mx, my, logBtnX, logBtnY, logBtnW, logBtnH) {
			g.openLog()
			return
		}
		if contains(mx, my, inventoryBtnX, inventoryBtnY, inventoryBtnW, inventoryBtnH) {
			g.ui.InventoryOpen = true
			g.ui.SettingsOpen = false
			g.ui.CodexOpen = false
			g.ui.LogOpen = false
			g.ui.NotificationHistoryOpen = false
			g.ui.InventoryHovered = ui.ToolNone
			return
		}
	}

	// Right-side panel: inventory picker and manual injection button.
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if openInvButtonHit(mx, my) {
			g.ui.InventoryOpen = true
			g.ui.SettingsOpen = false
			g.ui.CodexOpen = false
			g.ui.LogOpen = false
			g.ui.NotificationHistoryOpen = false
			g.ui.InventoryHovered = ui.ToolNone
			return
		}
		if injectButtonHit(mx, my) {
			g.triggerInjection()
			return
		}
		if g.handlePrestigePanelClick(mx, my) {
			return
		}
	}

	// Grid: place / reconfigure / erase.
	if pos, ok := cellAt(mx, my); ok {
		if inpututil.IsKeyJustPressed(ebiten.KeyH) {
			g.showComponentHelpAt(pos, mx, my)
			return
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
			input.PickToolAt(g.state, g.ui, pos)
			return
		}
		_, wheelY := ebiten.Wheel()
		if wheelY > 0 {
			input.ReconfigureBy(g.state, pos, 1)
			return
		}
		if wheelY < 0 {
			input.ReconfigureBy(g.state, pos, -1)
			return
		}
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			if g.ui.Selected != ui.ToolNone {
				input.PlaceFromTool(g.state, g.ui, pos)
			}
		} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
			input.Erase(g.state, pos)
		}
	}
}

func (g *Game) openSettings() {
	g.ui.SettingsOpen = true
	g.ui.CodexOpen = false
	g.ui.InventoryOpen = false
	g.ui.LogOpen = false
	g.ui.NotificationHistoryOpen = false
	g.ui.HotkeysOpen = false
	g.ui.PrestigeConfirmOpen = false
	g.ui.ResetArmed = false
	g.ui.LastSaveNotice = ""
}

func (g *Game) openCodex() {
	g.ui.CodexOpen = true
	g.ui.SettingsOpen = false
	g.ui.InventoryOpen = false
	g.ui.LogOpen = false
	g.ui.NotificationHistoryOpen = false
	g.ui.HotkeysOpen = false
	g.ui.CodexNotice = ""
	g.codexHovered = ""
	g.codexPinned = ""
}

func (g *Game) closeHelper() {
	g.ui.HelperOpen = false
	g.ui.HelperHeader = ""
	g.ui.HelperBody = ""
}

func (g *Game) toggleCodex() {
	if g.ui.CodexOpen {
		g.ui.CodexOpen = false
		g.ui.CodexNotice = ""
		g.codexHovered = ""
		g.codexPinned = ""
		return
	}
	g.openCodex()
}

func (g *Game) openLog() {
	g.ui.LogOpen = true
	g.ui.SettingsOpen = false
	g.ui.CodexOpen = false
	g.ui.InventoryOpen = false
	g.ui.NotificationHistoryOpen = false
	g.ui.HotkeysOpen = false
}

func (g *Game) toggleHotkeys() {
	if g.ui.HotkeysOpen {
		g.ui.HotkeysOpen = false
		return
	}
	g.ui.HotkeysOpen = true
	g.ui.SettingsOpen = false
	g.ui.CodexOpen = false
	g.ui.InventoryOpen = false
	g.ui.LogOpen = false
	g.ui.NotificationHistoryOpen = false
}

func (g *Game) handleHotkeysClick(mx, my int) {
	if contains(mx, my, hotkeysCloseX(), hotkeysCloseY(), closeBtnW, closeBtnH) || !hotkeysInPanel(mx, my) {
		g.ui.HotkeysOpen = false
	}
}

func (g *Game) toggleLog() {
	if g.ui.LogOpen {
		g.ui.LogOpen = false
		return
	}
	g.openLog()
}

func (g *Game) closeModals() {
	g.ui.SettingsOpen = false
	g.ui.CodexOpen = false
	g.ui.InventoryOpen = false
	g.ui.LogOpen = false
	g.ui.NotificationHistoryOpen = false
	g.ui.HotkeysOpen = false
	g.ui.ResetArmed = false
	g.ui.CodexNotice = ""
	g.ui.InventoryHovered = ui.ToolNone
	g.codexHovered = ""
	g.codexPinned = ""
}

func (g *Game) triggerInjection() {
	g.state.Inject()
}

func (g *Game) handleSettingsClick(mx, my int) {
	if contains(mx, my, saveBtnX(), saveBtnY(), saveBtnW, saveBtnH) {
		if g.save != nil {
			if err := g.save(); err == nil {
				g.ui.LastSaveNotice = "Saved"
				g.ui.AutosaveError = ""
			} else {
				log.Printf("save failed: %v", err)
				g.ui.LastSaveNotice = "Save failed: " + err.Error()
				g.ui.AutosaveError = err.Error()
			}
		}
		g.ui.ResetArmed = false
		return
	}
	if contains(mx, my, resetBtnX(), resetBtnY(), resetBtnW, resetBtnH) {
		if !g.ui.ResetArmed {
			g.ui.ResetArmed = true
			g.ui.LastSaveNotice = ""
			return
		}
		g.state.HardReset()
		g.income.reset(g.state.TickRate)
		g.ticksSinceSave = 0
		if g.reset != nil {
			if err := g.reset(); err == nil {
				g.ui.LastSaveNotice = "Reset"
				g.ui.AutosaveError = ""
			} else {
				log.Printf("reset save failed: %v", err)
				g.ui.LastSaveNotice = "Reset failed: " + err.Error()
				g.ui.AutosaveError = err.Error()
			}
		}
		g.ui.ResetArmed = false
		return
	}
	if contains(mx, my, trailsRowX(), trailsRowY(), trailsRowW, trailsRowH) {
		g.ui.TrailsEnabled = !g.ui.TrailsEnabled
		if !g.ui.TrailsEnabled {
			g.trail = g.trail[:0]
		}
		return
	}
	if contains(mx, my, historyBtnX(), historyBtnY(), historyBtnW, historyBtnH) {
		g.ui.NotificationHistoryOpen = true
		g.ui.NotificationHistoryScroll = 0
		g.ui.ResetArmed = false
		return
	}
	if contains(mx, my, hotkeysBtnX(), hotkeysBtnY(), hotkeysBtnW, hotkeysBtnH) {
		g.ui.SettingsOpen = false
		g.ui.HotkeysOpen = true
		g.ui.ResetArmed = false
		return
	}
	if contains(mx, my, closeBtnX(), closeBtnY(), closeBtnW, closeBtnH) {
		g.ui.SettingsOpen = false
		g.ui.ResetArmed = false
		return
	}
}

func (g *Game) handleNotificationHistoryClick(mx, my int) {
	if contains(mx, my, notifCloseX(), notifCloseY(), closeBtnW, closeBtnH) || !notifInPanel(mx, my) {
		g.ui.NotificationHistoryOpen = false
	}
}

func (g *Game) handleInventoryClick(mx, my int) {
	if contains(mx, my, invCloseX(), invCloseY(), closeBtnW, closeBtnH) {
		g.ui.InventoryOpen = false
		g.ui.InventoryHovered = ui.ToolNone
		return
	}
	if t := invToolAt(mx, my); t != ui.ToolNone {
		if !ui.IsToolUnlocked(g.state, t) {
			return
		}
		g.ui.Selected = t
		g.ui.InventoryOpen = false
		g.ui.InventoryHovered = ui.ToolNone
		return
	}
	if !invInPanel(mx, my) {
		g.ui.InventoryOpen = false
		g.ui.InventoryHovered = ui.ToolNone
	}
}

func (g *Game) handleLogClick(mx, my int) {
	if contains(mx, my, logCloseX(), logCloseY(), closeBtnW, closeBtnH) || !logInPanel(mx, my) {
		g.ui.LogOpen = false
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(colorBG)
	alpha := g.tickAlpha()
	g.updateTrail(alpha)
	drawGrid(screen, g.state, alpha, g.trail)
	drawPalette(screen, g.state, g.ui)
	drawHeader(screen, g.state, g.ui, g.income.perSecond())
	if g.ui.SettingsOpen {
		drawSettings(screen, g.ui)
	}
	if g.ui.CodexOpen {
		drawPeriodicTable(screen, g.state, g.ui, g.currentCodexFocus())
	}
	if g.ui.InventoryOpen {
		drawInventory(screen, g.state, g.ui)
	}
	if g.ui.LogOpen {
		drawCollectionLog(screen, g.state)
	}
	if g.ui.NotificationHistoryOpen {
		drawNotificationHistory(screen, g.state, g.ui)
	}
	if g.ui.HotkeysOpen {
		drawHotkeysModal(screen)
	}
	if g.ui.PrestigeConfirmOpen {
		drawPrestigeConfirmModal(screen, g.state)
	}
	if g.ui.HelperOpen {
		drawHelperModal(screen, g.ui)
	}
}

// tickAlpha returns the wall-clock fraction within the current sim tick, in
// [0, 1]. Clamps when Draw runs late (e.g. a slow frame backed up the cadence).
func (g *Game) tickAlpha() float64 {
	if g.tickDuration <= 0 {
		return 0
	}
	a := float64(time.Since(g.lastTickAt)) / float64(g.tickDuration)
	if a < 0 {
		return 0
	}
	if a > 1 {
		return 1
	}
	return a
}

func (g *Game) handleCodexClick(mx, my int) {
	if contains(mx, my, codexCloseX(), codexCloseY(), closeBtnW, closeBtnH) {
		g.ui.CodexOpen = false
		g.ui.CodexNotice = ""
		g.codexHovered = ""
		g.codexPinned = ""
		return
	}
	if e := g.currentCodexFocus(); e != "" {
		bx, by, bw, bh := codexUnlockButtonRect()
		if contains(mx, my, bx, by, bw, bh) && sim.IsElementUnlocked(g.state, e) && g.state.InjectionElement != e {
			g.state.InjectionElement = e
			g.ui.CodexNotice = "Injecting " + sim.ElementCatalog[e].Name
			return
		}
		if contains(mx, my, bx, by, bw, bh) && sim.IsElementPurchasable(g.state, e) {
			if err := sim.PurchaseElement(g.state, e); err != nil {
				g.ui.CodexNotice = "Unlock failed: " + err.Error()
				return
			}
			g.ui.CodexNotice = sim.ElementCatalog[e].Name + " unlocked"
			return
		}
	}
	if e, ok := codexElementAt(mx, my); ok {
		if g.codexPinned == e {
			g.codexPinned = ""
		} else {
			g.codexPinned = e
		}
		g.codexHovered = e
		return
	}
	if e := g.currentCodexFocus(); e != "" && contains(mx, my, codexCardX(), codexCardY(), codexCardW, codexCardH) {
		g.codexHovered = e
		return
	}
	if g.codexPinned != "" {
		g.codexPinned = ""
		g.codexHovered = ""
		return
	}
	if !contains(mx, my, codexPanelX(), codexPanelY(), codexPanelW(), codexPanelH()) {
		g.codexHovered = ""
	}
}

func (g *Game) updateCodexHover(mx, my int) {
	if e, ok := codexElementAt(mx, my); ok {
		g.codexHovered = e
		return
	}
	g.codexHovered = ""
}

func (g *Game) currentCodexFocus() sim.Element {
	return codexFocusedElement(g.codexHovered, g.codexPinned)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenW, screenH
}
