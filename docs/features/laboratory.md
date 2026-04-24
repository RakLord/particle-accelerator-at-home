# Laboratory

**Status:** Phase 4.

## Concept

The Laboratory is the post-prestige upgrade tree. It opens after the player's first prestige and houses upgrades paid in **Bond Points (BP)** — a meta-currency awarded by synthesising Bonds.

The Laboratory is the durable home of Bonds-derived progression. It is where the player spends the rewards of completing prestige loops to make future loops shorter, faster, or richer.

## Bond Points

Bond Points are awarded **once per Bond synthesised**, weighted by recipe complexity:

| Bond | BP awarded |
|---|---:|
| Methane | 1 |
| Acetylene | 1 |
| Ethylene | 2 |
| Benzene | 3 |
| Diamond | 3 |

Award is one-shot — re-prestiging does not re-award BP for already-owned Bonds. New Bonds (e.g. heavier-Element compounds when they ship) carry their own BP awards.

**MVP total: 10 BP.** Laboratory tree total cost: 31 BP. The player affords roughly ⅓ of the tree from MVP content alone; heavier-Element Bonds fund the rest.

`GameState.BondPoints int` tracks the remaining unspent balance. The total *earned* over the run history isn't tracked separately — only the spendable balance.

## Catalog

Numbers are first-draft and may shift after playtesting. The catalog lives in `internal/sim/laboratory.go`.

| Upgrade | Max purchases | Effect | Cost per purchase |
|---|---:|---|---:|
| Covalent Memory | 1 | Start each run with Helium already unlocked (skips $500 + research wall). | 1 |
| Stable Isotope | 1 | Per-Element research carries over at 30% on prestige. | 2 |
| Chain Reaction | 1 | Injectors start at 2× base emission rate (effective on Run 1 only — subsequent purchases supersede). | 2 |
| Carbon Core | 1 | Binder + Bonds tab unlock immediately on Run start, before reaching Carbon. | 1 |
| Dense Packing | 5 | All Binder per-Binder capacities ×2 per purchase (multiplicative; cap at 5 → ×32). | 1 / 2 / 3 / 4 / 5 |
| Auto-Inject Speed I | 1 | Auto-Inject cadence 10s → 8s. | 1 |
| Auto-Inject Speed II | 1 | 8s → 6s. | 2 |
| Auto-Inject Speed III | 1 | 6s → 5s. | 3 |
| Auto-Inject Speed IV | 1 | 5s → 4s. | 4 |

**Tree total: 31 BP.**

### Notes on individual upgrades

**Covalent Memory** + **Stable Isotope** under the hard reset.
The Genesis reset wipes both per-Element research *and* `UnlockedElements` (see `docs/features/prestige-genesis-ascension.md`). Both of these upgrades become much more impactful than a milder reset would imply. Their costs (1 BP, 2 BP) are deliberately low to make them attractive first purchases. Re-tune if playtesting shows them dominating the early Lab spend.

**Carbon Core.**
Skips the entire Carbon unlock chain on every future Run — no Boron grind, no $100K wall. Combined with `Stable Isotope`'s 30% research carryover, a Run 5+ player effectively jumps straight to Carbon banking on each prestige.

**Dense Packing.**
Multiplicative ×2 per purchase, applied to *per-Binder* base capacity (not the per-Element total). At level 5: a Carbon Binder holds 100 × 32 = 3,200 Subjects. Each level's cost climbs to keep the late tiers as long-tail goals.

**Auto-Inject Speed I-IV.**
See `docs/features/auto-injection.md`. Only the highest purchased level applies (`max(LaboratoryUpgrades["auto_inject_speed"], ...)`). The four entries together cost 10 BP — a substantial chunk of the tree, balancing the strong idle-mode payoff.

## Effect integration

Each Laboratory upgrade carries an `Apply(*GlobalModifiers, *GameState)` closure, identical in spirit to a `GlobalUpgrade` but with two arguments: most upgrades only need `*GlobalModifiers`, but a few (Stable Isotope, Carbon Core) need to read or write `*GameState` directly because they affect non-modifier state (research persistence, unlock flags).

The `rebuildModifiers(s)` pipeline (ADR 0010) is extended to invoke Laboratory upgrade closures alongside `PurchasedUpgrades` and `BondsState`. See ADR 0018 for the integration shape and the `*GameState` extension reasoning.

## Purchase flow

`PurchaseLabUpgrade(s *GameState, id LabUpgradeID) error`:

- Verifies upgrade exists and the player hasn't exceeded its `MaxPurchases`.
- Verifies `BondPoints >= cost`.
- Deducts BP, increments `LaboratoryUpgrades[id]`, calls `rebuildModifiers(s)`.

Tiered upgrades (Dense Packing, Auto-Inject Speed) use the same entry point — calling it with the same `id` advances the level by one until `MaxPurchases` is reached.

## Laboratory tab (UI)

Visible only after first prestige. Lives as a peer tab to Bonds in the right-side panel.

Layout:

- A **BP balance** at the top.
- A **list of upgrades**, each row showing: name, current level / max, next-purchase cost, effect description, Buy button (disabled if maxed or unaffordable).
- Tiered upgrades show their progress as a small bar (e.g. "Dense Packing: 2/5 — next: 3 BP").

Hovering shows extended description and the cumulative effect (e.g. for Dense Packing: "current: ×4 capacity → next: ×8 capacity").

## Save compatibility

`BondPoints int` and `LaboratoryUpgrades map[LabUpgradeID]int` are new fields with `omitempty`. Old saves load with empty values — no upgrades owned, zero BP. No save-envelope bump.

## Related

- `internal/sim/laboratory.go` — catalog, `PurchaseLabUpgrade`, integration with `rebuildModifiers`.
- `internal/sim/state.go` — `BondPoints`, `LaboratoryUpgrades`.
- `internal/ui/laboratory_tab.go` — tab UI.
- `docs/adr/0018-laboratory-upgrade-tree.md` — currency, catalog shape, modifier integration.
- `docs/adr/0010-global-modifier-pipeline.md` — pipeline that Laboratory upgrades extend.
- `docs/features/bonds-and-tokens.md` — where BP is earned.
- `docs/features/auto-injection.md` — what the cadence upgrades drive.
- `docs/features/prestige-genesis-ascension.md` — what the tree softens.
