package ui

type Tool int

const (
	ToolNone Tool = iota
	ToolInjector
	ToolAccelerator
	ToolMeshGrid
	ToolMagnetiser
	ToolElbow
	ToolCollector
	ToolErase
	// Phase 4 tools.
	ToolResonator
	ToolCatalyst
	ToolDuplicator
)

type UIState struct {
	Selected       Tool
	SettingsOpen   bool
	CodexOpen      bool
	InventoryOpen  bool
	LogOpen        bool
	SavePending    bool
	LastSaveNotice string
	CodexNotice    string
	ResetArmed     bool
	// AutosaveError is the most recent autosave error message, or empty if
	// the last autosave succeeded. Rendered in the header so silent failure
	// isn't possible.
	AutosaveError string
	// TrailsEnabled toggles the fading particle trail behind each Subject.
	// Session-scoped (not persisted); default on.
	TrailsEnabled bool
	// InventoryHovered tracks the card the cursor is over while the inventory
	// modal is open. Drives the right-hand description panel. Cleared on
	// modal close; not persisted.
	InventoryHovered Tool
}

func NewUIState() *UIState { return &UIState{Selected: ToolNone, TrailsEnabled: true} }
