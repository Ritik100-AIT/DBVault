package scheduler

import (
	"path/filepath"
	"testing"
	"time"
)

func TestSchedulerPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "schedules.json")
	s := NewSchedulerWithPath(path)

	sch := Schedule{
		ID:      "daily-backup",
		Cron:    "0 2 * * *",
		Name:    "daily-backup",
		NextRun: time.Now().Add(time.Hour),
	}

	if err := s.Add(sch); err != nil {
		t.Fatalf("failed to add schedule: %v", err)
	}

	schedules, err := s.List()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(schedules) != 1 {
		t.Fatalf("expected 1 schedule, got %d", len(schedules))
	}

	s2 := NewSchedulerWithPath(path)
	schedules, err = s2.List()
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	if len(schedules) != 1 || schedules[0].ID != "daily-backup" {
		t.Fatalf("expected persisted schedule, got %#v", schedules)
	}

	if err := s2.Remove("daily-backup"); err != nil {
		t.Fatalf("remove failed: %v", err)
	}

	schedules, err = s2.List()
	if err != nil {
		t.Fatalf("list after remove failed: %v", err)
	}
	if len(schedules) != 0 {
		t.Fatalf("expected no schedules, got %d", len(schedules))
	}
}
