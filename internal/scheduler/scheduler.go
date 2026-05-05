package scheduler

import "time"

type Schedule struct {
	ID      string
	Cron    string
	NextRun time.Time
}

type Scheduler struct {
}

func NewScheduler() *Scheduler {
	return &Scheduler{}
}

func (s *Scheduler) Add(schedule Schedule) error {
	return nil
}

func (s *Scheduler) List() ([]Schedule, error) {
	return nil, nil
}

func (s *Scheduler) Remove(id string) error {
	return nil
}
