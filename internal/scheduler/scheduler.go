package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"rail-go/internal/bot"
	"rail-go/internal/logger"
)

type Task struct {
	ID       string
	Interval time.Duration
	Execute  func() error
	Logger   *logger.Logger
}

type Scheduler struct {
	tasks  map[string]*Task
	mu     sync.RWMutex
	logger *logger.Logger
}

func New(logger *logger.Logger) *Scheduler {
	return &Scheduler{
		tasks:  make(map[string]*Task),
		logger: logger,
	}
}

func (s *Scheduler) ScheduleTask(ctx context.Context, task *Task) error {
	s.mu.Lock()
	s.tasks[task.ID] = task
	s.mu.Unlock()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("panic in scheduled task", "taskID", task.ID, "error", r)
			}
		}()

		ticker := time.NewTicker(task.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				s.logger.Info("shutting down scheduler", "taskID", task.ID)
				return
			case <-ticker.C:
				if err := task.Execute(); err != nil {
					s.logger.Error("task execution failed", "taskID", task.ID, "error", err)
				}
			}
		}
	}()

	return nil
}

func (s *Scheduler) RemoveTask(taskID string) {
	s.mu.Lock()
	delete(s.tasks, taskID)
	s.mu.Unlock()
}

func (s *Scheduler) GetTask(taskID string) (*Task, bool) {
	s.mu.RLock()
	task, exists := s.tasks[taskID]
	s.mu.RUnlock()
	return task, exists
}

var nowFunc = time.Now

func nextMonthlyRun() time.Time {
	now := nowFunc()

	year, month, _ := now.Date()
	loc := now.Location()

	// Calculate the last day of the current month.
	firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, loc)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	durationUntilEnd := lastOfMonth.Sub(now)
	var scheduledTime time.Time
	if durationUntilEnd.Hours() < 120.0 {
		daysLeft := int(durationUntilEnd.Hours() / 24)
		lastOfMonth = firstOfMonth.AddDate(0, 1, -daysLeft)
		scheduledTime = time.Date(year, month, lastOfMonth.Day(), 8, 0, 0, 0, loc)
	} else {
		scheduledTime = time.Date(year, month, lastOfMonth.Day()-5, 8, 0, 0, 0, loc)
	}

	return scheduledTime
}

func (s *Scheduler) ScheduleMonthlyTask(ctx context.Context, botService *bot.Service, chatID int64) error {
	task := &Task{
		ID:       "monthly_notification",
		Interval: s.calculateMonthlyInterval(),
		Execute: func() error {
			t := time.Now()
			return botService.SendMessage(ctx, chatID,
				fmt.Sprintf("%s need to use sibus", t.Local().Format("Mon Jan 2 15:04:05")))
		},
		Logger: s.logger,
	}

	return s.ScheduleTask(ctx, task)
}

func (s *Scheduler) calculateMonthlyInterval() time.Duration {
	nextRun := nextMonthlyRun()
	return time.Until(nextRun)
}
