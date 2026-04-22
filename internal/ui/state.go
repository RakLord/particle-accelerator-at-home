package ui

type Tool int

const (
	ToolNone Tool = iota
	ToolInjectorHydrogen
	ToolInjectorHelium
	ToolAccelerator
	ToolMeshGrid
	ToolMagnetiser
	ToolRotator
	ToolCollector
	ToolErase
)

type UIState struct {
	Selected          Tool
	SettingsOpen      bool
	CodexOpen         bool
	SavePending       bool
	LastSaveNotice    string
	CodexNotice       string
	ResetArmed        bool
	// AutosaveError is the most recent autosave error message, or empty if
	// the last autosave succeeded. Rendered in the header so silent failure
	// isn't possible.
	AutosaveError string
	// TrailsEnabled toggles the fading particle trail behind each Subject.
	// Session-scoped (not persisted); default on.
	TrailsEnabled bool
}

func NewUIState() *UIState { return &UIState{Selected: ToolNone, TrailsEnabled: true} }
