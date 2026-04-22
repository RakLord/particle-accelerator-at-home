package sim

// Injector is a source: it spawns Subjects every SpawnInterval ticks
// in its configured Direction. The spawn logic lives in Grid.Tick();
// Apply is a no-op so a Subject passing over an Injector is unaffected.
type Injector struct {
	Direction     Direction
	SpawnInterval int
	Element       Element
	TickCounter   int
}

func (*Injector) Kind() ComponentKind     { return KindInjector }
func (*Injector) Apply(s Subject) Subject { return s }
