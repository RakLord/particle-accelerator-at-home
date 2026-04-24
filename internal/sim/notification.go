package sim

// RecordNotification prepends a helper notification to the persisted history
// and caps the list to MaxNotificationLogEntries.
func (s *GameState) RecordNotification(header, body, timeHHMM string) {
	entry := NotificationEntry{
		Header:   header,
		Body:     body,
		TimeHHMM: timeHHMM,
		Tick:     s.Ticks,
	}
	s.NotificationLog = append([]NotificationEntry{entry}, s.NotificationLog...)
	s.trimNotificationLog()
}

func (s *GameState) trimNotificationLog() {
	if len(s.NotificationLog) > MaxNotificationLogEntries {
		s.NotificationLog = s.NotificationLog[:MaxNotificationLogEntries]
	}
}

// HasShownHelperMilestone reports whether a one-shot helper milestone has
// already been displayed for this save.
func (s *GameState) HasShownHelperMilestone(id string) bool {
	return id != "" && s.ShownHelperMilestones[id]
}

// MarkHelperMilestoneShown records a one-shot helper milestone as shown.
func (s *GameState) MarkHelperMilestoneShown(id string) {
	if id == "" {
		return
	}
	if s.ShownHelperMilestones == nil {
		s.ShownHelperMilestones = map[string]bool{}
	}
	s.ShownHelperMilestones[id] = true
}
