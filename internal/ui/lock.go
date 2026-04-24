package ui

import "particleaccelerator/internal/sim"

// IsToolUnlocked reports whether the player is allowed to select and place
// the given Tool in the current game state. Most tools are always unlocked;
// Element selection for Injectors is handled separately in the Codex.
//
// ToolNone reports unlocked (selecting nothing is always valid).
func IsToolUnlocked(s *sim.GameState, t Tool) bool {
	return true
}

// ToolLockReason returns a short user-facing explanation for why a Tool is
// locked, or "" if it's unlocked. Rendered in the inventory's description
// panel when the player hovers a greyed-out card.
func ToolLockReason(s *sim.GameState, t Tool) string {
	return ""
}

// KindForTool maps a Tool to the ComponentKind used for cost and inventory
// accounting. Returns "" for Tools that don't participate in the inventory
// system (ToolNone, ToolErase). Injector Element selection is free and lives in
// the Codex; only the component kind has a cost curve.
//
// This is the canonical mapping; render.palette and input both defer here
// to avoid drift.
func KindForTool(t Tool) sim.ComponentKind {
	switch t {
	case ToolInjector:
		return sim.KindInjector
	case ToolAccelerator:
		return sim.KindAccelerator
	case ToolMeshGrid:
		return sim.KindMeshGrid
	case ToolMagnetiser:
		return sim.KindMagnetiser
	case ToolElbow:
		return sim.KindRotator
	case ToolCollector:
		return sim.KindCollector
	case ToolResonator:
		return sim.KindResonator
	case ToolCatalyst:
		return sim.KindCatalyst
	case ToolDuplicator:
		return sim.KindDuplicator
	}
	return ""
}

// PlaceableTools is the ordered list of Tools shown in the inventory modal
// and summarised in the sidebar. ToolNone is excluded (it's "no selection");
// ToolErase is excluded (erasing has its own right-click affordance and
// doesn't belong in a stock/purchase grid).
var PlaceableTools = []Tool{
	ToolInjector,
	ToolAccelerator,
	ToolMeshGrid,
	ToolMagnetiser,
	ToolResonator,
	ToolCatalyst,
	ToolDuplicator,
	ToolElbow,
	ToolCollector,
}
