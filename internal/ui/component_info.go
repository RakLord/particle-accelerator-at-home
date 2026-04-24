package ui

// ToolInfo holds the user-facing copy for a placeable Tool. Name is the
// short label rendered on buttons; Tagline is a one-line summary for
// compact surfaces; Description is the full how-it-works text shown in
// the inventory's hover panel.
type ToolInfo struct {
	Name        string
	Tagline     string
	Description string
}

// ToolInfoCatalog is the single source of truth for component copy. Every
// Tool constant except ToolNone should have an entry — component_info_test.go
// asserts this.
var ToolInfoCatalog = map[Tool]ToolInfo{
	ToolInjector: {
		Name:    "Injector",
		Tagline: "Spawns the Codex-selected Element at a fixed cadence.",
		Description: "Emits a Subject every few ticks in its facing direction using the Element selected for injection in the Codex. " +
			"Spawns are blocked while the grid is at Max Load. " +
			"Hover it and use the scroll wheel to cycle the facing direction after placing.",
	},
	ToolAccelerator: {
		Name:    "Accelerator",
		Tagline: "Adds tier-driven Speed to a Subject passing through.",
		Description: "Adds +1 Speed at T1, scaling with each tier. " +
			"Directional — a Subject only accepts the boost when it enters along the Accelerator's orientation. " +
			"Stacks additively with every other Accelerator on the path.",
	},
	ToolMeshGrid: {
		Name:    "Mesh Grid",
		Tagline: "Throttles Speed by integer division. A tool, not a trap.",
		Description: "Divides the Subject's Speed by the tier divisor (÷2 at T1, ÷3 at T2, ÷4 at T3). " +
			"Only triggers when Speed is at or above the band floor, so slow Subjects pass through untouched rather than getting stuck at 0. " +
			"Use it to drop a fast Subject back into another component's activation band.",
	},
	ToolMagnetiser: {
		Name:    "Magnetiser",
		Tagline: "Adds tier-driven Magnetism — the second value axis.",
		Description: "Adds +1 Magnetism at T1, scaling with tier. " +
			"Magnetism multiplies into the collected-value formula alongside Mass and Speed, so chained Magnetisers build a parallel economy from the Speed-focused Accelerator line.",
	},
	ToolResonator: {
		Name:    "Resonator",
		Tagline: "Cluster-based Speed boost. Reward dense layouts.",
		Description: "Adds Speed to a passing Subject based on how many other Resonators sit in the four orthogonal neighbours. " +
			"Isolated Resonators are inert; a Resonator surrounded by four neighbours gives a big boost. " +
			"Tip: pack them tight — you're trading grid area for raw Speed.",
	},
	ToolCatalyst: {
		Name:    "Catalyst",
		Tagline: "Multiplies Mass, but only once research clears the threshold.",
		Description: "Inert until the passing Subject's Element has enough research (25 by default). " +
			"Once live, multiplies Mass by the tier factor (×1.5 at T1 → ×3.0 at T3). " +
			"Stacks multiplicatively across multiple Catalysts on the same path — a late-game payoff for investing in research.",
	},
	ToolDuplicator: {
		Name:    "Duplicator",
		Tagline: "T-junction that emits two Subjects from one.",
		Description: "A Subject entering the input side is consumed; two Subjects leave along the perpendicular output sides, each carrying a fraction of the original Mass (×0.5 per output at T1, rising with tier). " +
			"Speed and Magnetism copy unchanged. " +
			"At T1 total Mass is conserved — the win comes from parallelism. Higher tiers actively create Mass.",
	},
	ToolElbow: {
		Name:    "Elbow (turn)",
		Tagline: "Redirects a Subject by 90°.",
		Description: "Bends a Subject's direction into its configured turn. " +
			"Left-click on an existing Elbow to cycle its orientation. " +
			"Use these to route long paths through your Accelerator and Magnetiser chains before hitting the Collector.",
	},
	ToolCollector: {
		Name:    "Collector (endpoint)",
		Tagline: "Removes the Subject and pays out $USD + research.",
		Description: "Endpoint cell. Consumes the Subject on arrival, credits $USD based on its Mass, Speed, Magnetism, and the Element's research multiplier, and advances that Element's research level. " +
			"Every path needs one — without a Collector, Subjects loop until Max Load throttles further spawns.",
	},
	ToolErase: {
		Name:    "Erase",
		Tagline: "Remove whatever is on a cell and return it to inventory.",
		Description: "Click a cell to remove its component or Collector. " +
			"The removed item returns to your available inventory — placement is free, removal is free, only purchasing new components costs $USD.",
	},
}
