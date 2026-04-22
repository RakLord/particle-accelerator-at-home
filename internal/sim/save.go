package sim

import (
	"encoding/json"
	"fmt"

	"particleaccelerator/internal/save"
)

const (
	saveKey        = "state"
	currentVersion = 1
)

type saveEnvelope struct {
	Version int             `json:"version"`
	State   json.RawMessage `json:"state"`
}

type cellJSON struct {
	IsCollector bool            `json:"is_collector,omitempty"`
	Kind        ComponentKind   `json:"kind,omitempty"`
	Component   json.RawMessage `json:"component,omitempty"`
}

func (c Cell) MarshalJSON() ([]byte, error) {
	out := cellJSON{IsCollector: c.IsCollector}
	if c.Component != nil {
		inner, err := json.Marshal(c.Component)
		if err != nil {
			return nil, err
		}
		out.Kind = c.Component.Kind()
		out.Component = inner
	}
	return json.Marshal(out)
}

func (c *Cell) UnmarshalJSON(data []byte) error {
	var in cellJSON
	if err := json.Unmarshal(data, &in); err != nil {
		return err
	}
	c.IsCollector = in.IsCollector
	if in.Kind == "" || len(in.Component) == 0 {
		c.Component = nil
		return nil
	}
	comp, err := newComponentByKind(in.Kind)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(in.Component, comp); err != nil {
		return err
	}
	c.Component = comp
	return nil
}

func newComponentByKind(kind ComponentKind) (Component, error) {
	switch kind {
	case KindInjector:
		return &Injector{}, nil
	case KindAccelerator:
		return &SimpleAccelerator{}, nil
	case KindRotator:
		return &Rotator{}, nil
	}
	return nil, fmt.Errorf("sim: unknown component kind %q", kind)
}

// Save serializes the current GameState into the versioned envelope and
// writes it through internal/save.
func (s *GameState) Save() error {
	body, err := json.Marshal(s)
	if err != nil {
		return err
	}
	env := saveEnvelope{Version: currentVersion, State: body}
	blob, err := json.Marshal(env)
	if err != nil {
		return err
	}
	return save.Write(saveKey, string(blob))
}

// Load reads the versioned envelope and returns the restored GameState.
// If no save exists, returns (nil, false, nil). Unknown versions return an
// error so the caller can decide to boot with default state.
func Load() (*GameState, bool, error) {
	raw, ok, err := save.Read(saveKey)
	if err != nil {
		return nil, false, err
	}
	if !ok || raw == "" {
		return nil, false, nil
	}
	var env saveEnvelope
	if err := json.Unmarshal([]byte(raw), &env); err != nil {
		return nil, false, err
	}
	if env.Version != currentVersion {
		return nil, false, fmt.Errorf("sim: unsupported save version %d", env.Version)
	}
	state := NewGameState()
	if err := json.Unmarshal(env.State, state); err != nil {
		return nil, false, err
	}
	if state.Grid == nil {
		state.Grid = NewGrid()
	}
	if state.Research == nil {
		state.Research = map[Element]int{}
	}
	// Recompute CurrentLoad from subjects-in-flight to avoid drift.
	state.CurrentLoad = 0
	for _, sub := range state.Grid.Subjects {
		state.CurrentLoad += sub.Load
	}
	return state, true, nil
}
