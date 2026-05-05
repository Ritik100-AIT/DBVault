package scheduler

import (
	"fmt"
	"time"
)

type Schedule struct {
	ID      string
	Cron    string
	NextRun time.Time
	Name    string
}

type Scheduler struct {
	schedules []Schedule
}

func NewScheduler() *Scheduler {
	return &Scheduler{schedules: []Schedule{}}
}

func (s *Scheduler) Add(schedule Schedule) error {
	s.schedules = append(s.schedules, schedule)
	fmt.Printf("Added schedule: %s\n", schedule.Name)
	return nil
}

func (s *Scheduler) List() ([]Schedule, error) {
	return s.schedules, nil
}

func (s *Scheduler) Remove(id string) error {
	for i, sch := range s.schedules {
		if sch.ID == id {
			s.schedules = append(s.schedules[:i], s.schedules[i+1:]...)
			fmt.Printf("Removed schedule: %s\n", id)
			return nil
		}
	}
	return fmt.Errorf("schedule not found: %s", id)
}
