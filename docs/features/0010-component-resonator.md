# Resonator

**Status:** Phase 3.

## Behaviour

A Resonator transforms a Subject that enters its cell by adding to its **Speed** based on the number of other Resonators in the four orthogonally-adjacent cells. Zero neighbours is inert; one neighbour is a modest boost; four neighbours is a strong boost.

A Subject passing through an isolated Resonator receives no benefit. The component exists to reward **clusters**.

### Adjacency

Only N/S/E/W neighbours count. Diagonals do not. The Resonator on the centre of a `+` of five Resonators gets the maximum bonus.

Resonators do not chain-multiply. A Subject traversing three connected Resonators in a row receives each cell's contribution independently — Resonator A's bonus is based on A's neighbours, Resonator B's bonus is based on B's neighbours, and so on. This is intentional: chain-multiplication would explode the value space as grids grow.

## Design intent

Resonator is the first component whose gameplay depends on **spatial layout** rather than pure path sequencing. Players who discover that packing Resonators reshapes the economy of grid area — dense clusters give huge bonuses but consume scarce cells that can't also hold Accelerators or Magnetisers. Build diversity, not just build length.

The gameplay pairs naturally with Mesh Grid in the same build: throttle a fast Subject down into a Resonator cluster to stack speed back up in a controlled way.

## Tiers

Tierable. See `docs/features/0011-component-tiers.md`. Higher tiers increase the per-neighbour Speed contribution.

| Tier | Speed bonus per neighbour |
|---|---|
| T1 | `+1` |
| T2 | `+2` |
| T3 | `+3` |

Tier tables live in `internal/sim/components/resonator.go`.

## Related

- `internal/sim/components/resonator.go`
- `docs/adr/0008-apply-context-and-grid-view.md` — grid read is what makes this component possible.
- `docs/features/0011-component-tiers.md`
- `docs/features/0001-value-formula.md` — Speed feeds collected value.
