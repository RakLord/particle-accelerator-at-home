package sim

// Layer identifies which reset layer of the game the GameState is in. Layers
// are the game's prestige axis: each layer has its own Elements, Components,
// and currency context; ascending to a higher layer resets the current one
// and awards a meta-currency that carries across.
//
// Phase 2 only ships Genesis. Future layers add constants here. The type is
// a string so the serialized form is stable across name changes in code.
type Layer string

const (
	// LayerGenesis is the base reset layer — the game as shipped today.
	LayerGenesis Layer = "genesis"
)
