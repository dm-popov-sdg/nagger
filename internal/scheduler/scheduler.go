package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dm-popov-sdg/nagger/internal/types"
)

// TaskSender defines the interface for sending tasks
type TaskSender interface {
	SendDailyReminder(ctx context.Context, chatID int64, tasks []string) error
	SendDailyReminderWithTasks(ctx context.Context, chatID int64, tasks []types.TaskWithID) error
}

// TaskGetter defines the interface for getting tasks
type TaskGetter interface {
	GetAllActiveTasks(ctx context.Context) (map[int64][]Task, error)
}

// Task represents a task (simplified interface)
type Task interface {
	GetDescription() string
	GetID() string
	GetStatus() string
}

// Scheduler handles periodic task reminders
type Scheduler struct {
	storage      TaskGetter
	bot          TaskSender
	reminderTime string
	timezone     *time.Location
	stopChan     chan struct{}
}

// NewScheduler creates a new scheduler instance
func NewScheduler(storage TaskGetter, bot TaskSender, reminderTime, timezone string) (*Scheduler, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone %s: %w", timezone, err)
	}

	return &Scheduler{
		storage:      storage,
		bot:          bot,
		reminderTime: reminderTime,
		timezone:     loc,
		stopChan:     make(chan struct{}),
	}, nil
}

// Start begins the scheduler
func (s *Scheduler) Start(ctx context.Context) {
	go s.run(ctx)
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	close(s.stopChan)
}

func (s *Scheduler) run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Printf("Scheduler started. Will send reminders at %s %s", s.reminderTime, s.timezone)

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			if s.shouldSendReminder() {
				s.sendReminders(ctx)
			}
		}
	}
}

func (s *Scheduler) shouldSendReminder() bool {
	now := time.Now().In(s.timezone)
	currentTime := now.Format("15:04")
	return currentTime == s.reminderTime
}

func (s *Scheduler) sendReminders(ctx context.Context) {
	log.Println("Sending daily reminders...")

	tasks, err := s.storage.GetAllActiveTasks(ctx)
	if err != nil {
		log.Printf("Error getting tasks: %v", err)
		return
	}

	for chatID, chatTasks := range tasks {
		if len(chatTasks) == 0 {
			continue
		}

		// Convert to types.TaskWithID interface
		taskInterfaces := make([]types.TaskWithID, len(chatTasks))
		for i, task := range chatTasks {
			taskInterfaces[i] = task
		}

		// Send reminder with interactive task list
		if err := s.bot.SendDailyReminderWithTasks(ctx, chatID, taskInterfaces); err != nil {
			log.Printf("Error sending reminder to chat %d: %v", chatID, err)
		}
	}

	log.Println("Daily reminders sent")
}
