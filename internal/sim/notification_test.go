package sim

import "testing"

func TestRecordNotificationCapsNewestFirst(t *testing.T) {
	s := NewGameState()
	for i := 0; i < MaxNotificationLogEntries+5; i++ {
		s.Ticks = uint64(i)
		s.RecordNotification("Header "+string(rune('A'+i)), "Body", "12:34")
	}

	if got := len(s.NotificationLog); got != MaxNotificationLogEntries {
		t.Fatalf("NotificationLog length: got %d want %d", got, MaxNotificationLogEntries)
	}
	if got := s.NotificationLog[0].Tick; got != MaxNotificationLogEntries+4 {
		t.Fatalf("newest tick: got %d want %d", got, MaxNotificationLogEntries+4)
	}
	if got := s.NotificationLog[len(s.NotificationLog)-1].Tick; got != 5 {
		t.Fatalf("oldest retained tick: got %d want 5", got)
	}
}

func TestHelperMilestoneShownMap(t *testing.T) {
	s := NewGameState()
	if s.HasShownHelperMilestone("first-five-usd") {
		t.Fatalf("new save should not have milestone marked")
	}
	s.MarkHelperMilestoneShown("first-five-usd")
	if !s.HasShownHelperMilestone("first-five-usd") {
		t.Fatalf("milestone should be marked shown")
	}
	s.MarkHelperMilestoneShown("")
	if s.HasShownHelperMilestone("") {
		t.Fatalf("empty milestone IDs should be ignored")
	}
}
