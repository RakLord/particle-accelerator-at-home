# Collection Log

**Status:** Phase 3.

The Collection Log is a header tab that opens a modal showing the most recent 10 collected Subjects. It is a quick audit surface for understanding why recent payouts were high or low.

## Contents

Rows are newest first. Each row shows the Subject stats that feed collected value:

- collection tick
- Element
- Mass
- Speed
- Magnetism
- pre-collection research level
- awarded $USD value

## State

`GameState.CollectionLog` stores up to `sim.MaxCollectionLogEntries` entries and is persisted as additive save data. Entries are recorded at collection time from the actual payout path, before research increments, so the displayed research level matches the multiplier used for that payout.
