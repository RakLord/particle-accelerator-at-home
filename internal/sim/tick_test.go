package sim

import (
	"testing"

	"particleaccelerator/internal/bignum"
)

func TestTickMovesSubject(t *testing.T) {
	s := NewGameState()
	s.Grid.Cells[2][2].Component = &testPipe{}
	s.Grid.Subjects = append(s.Grid.Subjects, Subject{
		Element:     ElementHydrogen,
		Mass:        bignum.One(),
		Speed:       SpeedFromInt(SpeedDivisor), // one cell per tick
		Direction:   DirEast,
		InDirection: DirEast,
		Position:    Position{X: 1, Y: 2},
		Load:        1,
	})
	s.CurrentLoad = 1
	s.Tick()
	if len(s.Grid.Subjects) != 1 {
		t.Fatalf("expected 1 subject, got %d", len(s.Grid.Subjects))
	}
	got := s.Grid.Subjects[0].Position
	if got != (Position{X: 2, Y: 2}) {
		t.Fatalf("expected position (2,2), got %v", got)
	}
}

func TestCollectorAwardsUSD(t *testing.T) {
	s := NewGameState()
	s.Grid.Cells[2][2].IsCollector = true
	s.Grid.Subjects = append(s.Grid.Subjects, Subject{
		Element:   ElementHydrogen,
		Mass:      bignum.FromInt(2),
		Speed:     SpeedFromInt(SpeedDivisor),
		Direction: DirEast,
		Position:  Position{X: 1, Y: 2},
		Load:      1,
	})
	s.CurrentLoad = 1
	s.Tick()
	if len(s.Grid.Subjects) != 0 {
		t.Fatalf("expected subject removed after collection, got %d", len(s.Grid.Subjects))
	}
	// collectValue = Mass * Speed * speedK = 2 * SpeedDivisor * 1 (Hydrogen mult 1.0, research 0).
	wantUSD := bignum.FromInt(2 * SpeedDivisor)
	if !s.USD.Eq(wantUSD) {
		t.Fatalf("expected USD %v, got %v", wantUSD, s.USD)
	}
	if s.Research[ElementHydrogen] != 1 {
		t.Fatalf("expected research 1, got %d", s.Research[ElementHydrogen])
	}
	stats := s.BestStats[ElementHydrogen]
	if stats.MaxSpeed != SpeedFromInt(SpeedDivisor) {
		t.Fatalf("expected MaxSpeed %d, got %d", SpeedFromInt(SpeedDivisor), stats.MaxSpeed)
	}
	if !stats.MaxMass.Eq(bignum.FromInt(2)) {
		t.Fatalf("expected MaxMass 2, got %v", stats.MaxMass)
	}
	if !stats.MaxCollectedValue.Eq(wantUSD) {
		t.Fatalf("expected MaxCollectedValue %v, got %v", wantUSD, stats.MaxCollectedValue)
	}
	if s.CurrentLoad != 0 {
		t.Fatalf("expected CurrentLoad 0, got %d", s.CurrentLoad)
	}
	if len(s.CollectionLog) != 1 {
		t.Fatalf("expected one collection log entry, got %d", len(s.CollectionLog))
	}
	entry := s.CollectionLog[0]
	if entry.Element != ElementHydrogen || !entry.Mass.Eq(bignum.FromInt(2)) || entry.Speed != SpeedFromInt(SpeedDivisor) {
		t.Fatalf("unexpected log entry subject stats: %#v", entry)
	}
	if entry.ResearchLevel != 0 {
		t.Fatalf("log ResearchLevel = %d, want pre-increment 0", entry.ResearchLevel)
	}
	if !entry.Value.Eq(wantUSD) {
		t.Fatalf("log Value = %v, want %v", entry.Value, wantUSD)
	}
}

func TestCollectionLogKeepsRecentTenNewestFirst(t *testing.T) {
	s := NewGameState()
	for i := 0; i < MaxCollectionLogEntries+2; i++ {
		s.Ticks = uint64(i)
		s.recordCollectionLog(Subject{
			Element:   ElementHydrogen,
			Mass:      bignum.FromInt(i + 1),
			Speed:     SpeedFromInt(i + 1),
			Magnetism: bignum.FromInt(i),
		}, i, bignum.FromInt((i+1)*100))
	}

	if len(s.CollectionLog) != MaxCollectionLogEntries {
		t.Fatalf("log length = %d, want %d", len(s.CollectionLog), MaxCollectionLogEntries)
	}
	if got := s.CollectionLog[0].Tick; got != uint64(MaxCollectionLogEntries+1) {
		t.Fatalf("newest tick = %d, want %d", got, MaxCollectionLogEntries+1)
	}
	if got := s.CollectionLog[len(s.CollectionLog)-1].Tick; got != 2 {
		t.Fatalf("oldest retained tick = %d, want 2", got)
	}
}

func TestBestStatsUpdateOnCollectionOnly(t *testing.T) {
	s := NewGameState()
	s.Grid.Cells[2][2].IsCollector = true
	s.Grid.Subjects = append(s.Grid.Subjects,
		Subject{
			Element:   ElementHydrogen,
			Mass:      bignum.FromInt(2),
			Speed:     SpeedFromInt(SpeedDivisor),
			Direction: DirEast,
			Position:  Position{X: 1, Y: 2},
			Load:      1,
		},
		Subject{
			Element:   ElementHydrogen,
			Mass:      bignum.FromInt(9),
			Speed:     SpeedFromInt(SpeedDivisor * 2),
			Direction: DirEast,
			Position:  Position{X: 0, Y: 0},
			Load:      1,
		},
	)
	s.CurrentLoad = 2

	s.Tick()

	stats := s.BestStats[ElementHydrogen]
	if stats.MaxSpeed != SpeedFromInt(SpeedDivisor) {
		t.Fatalf("off-grid subject should not update MaxSpeed: got %d want %d", stats.MaxSpeed, SpeedFromInt(SpeedDivisor))
	}
	if !stats.MaxMass.Eq(bignum.FromInt(2)) {
		t.Fatalf("off-grid subject should not update MaxMass: got %v want 2", stats.MaxMass)
	}
}

func TestSubjectLeavingPipeNetworkIsDestroyed(t *testing.T) {
	s := NewGameState()
	// Starting cell has pipe; the next cell east is empty, so stepping into it
	// leaves the network.
	s.Grid.Cells[1][1].Component = &testPipe{}
	s.Grid.Subjects = append(s.Grid.Subjects, Subject{
		Element:     ElementHydrogen,
		Speed:       SpeedFromInt(SpeedDivisor), // one cell per tick
		Direction:   DirEast,
		InDirection: DirEast,
		Position:    Position{X: 1, Y: 1},
		Load:        1,
	})
	s.CurrentLoad = 1
	s.Tick()
	if len(s.Grid.Subjects) != 0 {
		t.Fatalf("expected subject destroyed on empty cell entry, got %d", len(s.Grid.Subjects))
	}
	if s.CurrentLoad != 0 {
		t.Fatalf("expected CurrentLoad 0 after destruction, got %d", s.CurrentLoad)
	}
}

func TestSubjectOffGridIsRemoved(t *testing.T) {
	s := NewGameState()
	s.Grid.Subjects = append(s.Grid.Subjects, Subject{
		Element:   ElementHydrogen,
		Speed:     SpeedFromInt(SpeedDivisor),
		Direction: DirEast,
		Position:  Position{X: GridSize - 1, Y: 0},
		Load:      1,
	})
	s.CurrentLoad = 1
	s.Tick()
	if len(s.Grid.Subjects) != 0 {
		t.Fatalf("expected subject removed after falling off grid")
	}
	if s.CurrentLoad != 0 {
		t.Fatalf("expected CurrentLoad 0, got %d", s.CurrentLoad)
	}
}

// A base-Speed=1 Subject moves exactly once per SpeedDivisor ticks.
func TestBaseSpeedAdvancesEverySpeedDivisorTicks(t *testing.T) {
	s := NewGameState()
	s.Grid.Cells[0][1].Component = &testPipe{}
	s.Grid.Subjects = append(s.Grid.Subjects, Subject{
		Element:     ElementHydrogen,
		Speed:       SpeedFromInt(1),
		Direction:   DirEast,
		InDirection: DirEast,
		Position:    Position{X: 0, Y: 0},
		Load:        1,
	})
	s.CurrentLoad = 1

	// For SpeedDivisor-1 ticks the Subject should stay put (progress accumulates).
	for i := 1; i < SpeedDivisor; i++ {
		s.Tick()
		if got := s.Grid.Subjects[0].Position; got != (Position{X: 0, Y: 0}) {
			t.Fatalf("tick %d: expected still at (0,0), got %v", i, got)
		}
	}
	// The SpeedDivisor-th tick crosses one cell.
	s.Tick()
	if got := s.Grid.Subjects[0].Position; got != (Position{X: 1, Y: 0}) {
		t.Fatalf("after %d ticks: expected (1,0), got %v", SpeedDivisor, got)
	}
	if got := s.Grid.Subjects[0].StepProgress; got != 0 {
		t.Fatalf("StepProgress after crossing should be 0, got %d", got)
	}
}

// Speed=2 moves once per SpeedDivisor/2 = 5 ticks.
func TestDoubleSpeedAdvancesHalfAsOften(t *testing.T) {
	s := NewGameState()
	s.Grid.Cells[0][1].Component = &testPipe{}
	s.Grid.Cells[0][2].Component = &testPipe{}
	s.Grid.Subjects = append(s.Grid.Subjects, Subject{
		Element:     ElementHydrogen,
		Speed:       SpeedFromInt(2),
		Direction:   DirEast,
		InDirection: DirEast,
		Position:    Position{X: 0, Y: 0},
		Load:        1,
	})
	s.CurrentLoad = 1

	// After 5 ticks, expect one cell crossed.
	for range SpeedDivisor / 2 {
		s.Tick()
	}
	if got := s.Grid.Subjects[0].Position; got != (Position{X: 1, Y: 0}) {
		t.Fatalf("expected (1,0) after 5 ticks, got %v", got)
	}
	// After 5 more ticks, a second cell.
	for range SpeedDivisor / 2 {
		s.Tick()
	}
	if got := s.Grid.Subjects[0].Position; got != (Position{X: 2, Y: 0}) {
		t.Fatalf("expected (2,0) after 10 ticks, got %v", got)
	}
}

// A single tick that crosses multiple cells (Speed ≥ 2·SpeedDivisor) records
// all entered cells in Path, including elbow turns.
func TestTickPathRecordsAcrossRotator(t *testing.T) {
	s := NewGameState()
	// Place a north-facing elbow in the middle of the row.
	// Note: sim package can't import components, so we use a minimal inline
	// stand-in via a closure component below.
	s.Grid.Cells[1][1].Component = &testPipe{}
	s.Grid.Cells[1][2].Component = &testRightTurn{}
	s.Grid.Cells[0][2].Component = &testPipe{}
	s.Grid.Subjects = append(s.Grid.Subjects, Subject{
		Element:     ElementHydrogen,
		Speed:       SpeedFromInt(3 * SpeedDivisor), // cross three cells in one tick
		Direction:   DirEast,
		InDirection: DirEast,
		Position:    Position{X: 0, Y: 1},
		Load:        1,
	})
	s.CurrentLoad = 1
	s.Tick()

	sub := s.Grid.Subjects[0]
	wantPath := []Position{
		{X: 0, Y: 1},
		{X: 1, Y: 1},
		{X: 2, Y: 1}, // elbow turns East -> North here
		{X: 2, Y: 0},
	}
	if len(sub.Path) != len(wantPath) {
		t.Fatalf("Path length: got %d want %d (%v)", len(sub.Path), len(wantPath), sub.Path)
	}
	for i, p := range wantPath {
		if sub.Path[i] != p {
			t.Fatalf("Path[%d]: got %v want %v", i, sub.Path[i], p)
		}
	}
	if sub.Direction != DirNorth {
		t.Fatalf("Direction after elbow: got %v want DirNorth", sub.Direction)
	}
	if sub.Position != (Position{X: 2, Y: 0}) {
		t.Fatalf("Position after tick: got %v want (2,0)", sub.Position)
	}
}

// testRightTurn is a minimal in-sim elbow used by path tests, avoiding an
// import cycle with the components package.
type testRightTurn struct{}

func (*testRightTurn) Kind() ComponentKind { return ComponentKind("test_right_turn") }
func (*testRightTurn) Apply(_ ApplyContext, s Subject) (Subject, bool) {
	s.Direction = DirNorth
	return s, false
}

// testPipe is an in-sim no-op Component. Tests use it to "lay pipe" on cells
// so Subjects can traverse them without tripping the empty-cell destroy rule.
type testPipe struct{}

func (*testPipe) Kind() ComponentKind                             { return ComponentKind("test_pipe") }
func (*testPipe) Apply(_ ApplyContext, s Subject) (Subject, bool) { return s, false }
