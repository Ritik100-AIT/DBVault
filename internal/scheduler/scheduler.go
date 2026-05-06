package scheduler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Schedule represents a configured backup schedule.
type Schedule struct {
	ID      string    `json:"id"`
	Cron    string    `json:"cron"`
	NextRun time.Time `json:"next_run"`
	Name    string    `json:"name"`
}

type Scheduler struct {
	mu        sync.Mutex
	schedules []Schedule
	storePath string
}

// NewScheduler creates a Scheduler that persists schedules under ~/.dbvault/schedules.json.
func NewScheduler() *Scheduler {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	path := filepath.Join(home, ".dbvault", "schedules.json")
	return NewSchedulerWithPath(path)
}

// NewSchedulerWithPath creates a Scheduler backed by a specific JSON file path.
func NewSchedulerWithPath(path string) *Scheduler {
	s := &Scheduler{storePath: path}
	_ = s.load()
	return s
}

// Add stores a schedule and persists the updated schedule list.
func (s *Scheduler) Add(schedule Schedule) error {
	if schedule.ID == "" {
		return fmt.Errorf("schedule id is required")
	}
	if schedule.Cron == "" {
		return fmt.Errorf("cron expression is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.schedules = append(s.schedules, schedule)
	return s.save()
}

// List returns the current schedule registry.
func (s *Scheduler) List() ([]Schedule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return append([]Schedule(nil), s.schedules...), nil
}

// Remove deletes a schedule by its ID and persists the change.
func (s *Scheduler) Remove(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, sch := range s.schedules {
		if sch.ID == id {
			s.schedules = append(s.schedules[:i], s.schedules[i+1:]...)
			return s.save()
		}
	}
	return fmt.Errorf("schedule not found: %s", id)
}

// save writes the schedule registry to disk as JSON.
func (s *Scheduler) save() error {
	if err := os.MkdirAll(filepath.Dir(s.storePath), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.schedules, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.storePath, data, 0o600)
}

// load reads the persisted schedule registry from disk.
func (s *Scheduler) load() error {
	data, err := os.ReadFile(s.storePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &s.schedules)
}
