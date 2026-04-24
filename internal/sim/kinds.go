package sim

// ComponentKind identifiers for every placeable cell type.
//
// These live in the sim package (not internal/sim/components) so that
// sim-level code — in particular the component cost catalog in
// component_cost.go — can reference them without creating an import
// cycle (the components package already imports sim).
const (
	KindInjector    ComponentKind = "injector"
	KindAccelerator ComponentKind = "accelerator"
	KindMeshGrid    ComponentKind = "mesh_grid"
	KindMagnetiser  ComponentKind = "magnetiser"
	KindRotator     ComponentKind = "rotator"
	KindPipe        ComponentKind = "pipe"

	// Phase 4 additions. See docs/features/0008-component-catalyst.md, 0009-component-duplicator.md, 0010-component-resonator.md, 0019-component-compressor.md.
	KindResonator  ComponentKind = "resonator"
	KindCatalyst   ComponentKind = "catalyst"
	KindDuplicator ComponentKind = "duplicator"
	KindCompressor ComponentKind = "compressor"

	// KindCollector is used only for cost/ownership accounting. Collectors
	// are stored as cell.IsCollector (see Cell), not Component instances,
	// and are intentionally NOT registered with componentRegistry. A future
	// refactor making Collector a real Component would drop the IsCollector
	// bool and also drop this special case — see
	// docs/adr/0005-component-cost-and-inventory.md.
	KindCollector ComponentKind = "collector"
)
