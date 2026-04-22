package ui

type Tool int

const (
	ToolNone Tool = iota
	ToolInjector
	ToolAccelerator
	ToolRotator
	ToolCollector
	ToolErase
)

type UIState struct {
	Selected       Tool
	SettingsOpen   bool
	SavePending    bool
	LastSaveNotice string
	ResetArmed     bool
	// AutosaveError is the most recent autosave error message, or empty if
	// the last autosave succeeded. Rendered in the header so silent failure
	// isn't possible.
	AutosaveError string
}

func NewUIState() *UIState { return &UIState{Selected: ToolNone} }
