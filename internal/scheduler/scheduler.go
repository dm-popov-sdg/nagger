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

// SettingsGetter defines the interface for getting user settings
type SettingsGetter interface {
	GetUserSettings(ctx context.Context, chatID int64) (*UserSettings, error)
	GetAllUserSettings(ctx context.Context) (map[int64]*UserSettings, error)
}

// UserSettings represents user-specific settings
type UserSettings struct {
	ChatID       int64
	ReminderTime string
	Timezone     string
}

// Task represents a task (simplified interface)
type Task interface {
	GetDescription() string
	GetID() string
	GetStatus() string
}

// Scheduler handles periodic task reminders
type Scheduler struct {
	storage         TaskGetter
	settingsStorage SettingsGetter
	bot             TaskSender
	defaultTime     string
	defaultTimezone *time.Location
	stopChan        chan struct{}
}

// NewScheduler creates a new scheduler instance
func NewScheduler(storage TaskGetter, settingsStorage SettingsGetter, bot TaskSender, defaultTime, defaultTimezone string) (*Scheduler, error) {
	loc, err := time.LoadLocation(defaultTimezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone %s: %w", defaultTimezone, err)
	}

	return &Scheduler{
		storage:         storage,
		settingsStorage: settingsStorage,
		bot:             bot,
		defaultTime:     defaultTime,
		defaultTimezone: loc,
		stopChan:        make(chan struct{}),
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

	log.Printf("Scheduler started. Default reminder time: %s %s", s.defaultTime, s.defaultTimezone)

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.sendReminders(ctx)
		}
	}
}

func (s *Scheduler) shouldSendReminderForUser(reminderTime, timezone string) bool {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.Printf("Invalid timezone %s, using default: %v", timezone, err)
		loc = s.defaultTimezone
	}

	now := time.Now().In(loc)
	currentTime := now.Format("15:04")
	return currentTime == reminderTime
}

func (s *Scheduler) sendReminders(ctx context.Context) {
	tasks, err := s.storage.GetAllActiveTasks(ctx)
	if err != nil {
		log.Printf("Error getting tasks: %v", err)
		return
	}

	// Get all user settings
	userSettings, err := s.settingsStorage.GetAllUserSettings(ctx)
	if err != nil {
		log.Printf("Error getting user settings: %v", err)
		// Continue with default settings
		userSettings = make(map[int64]*UserSettings)
	}

	// Check each chat with tasks
	for chatID, chatTasks := range tasks {
		if len(chatTasks) == 0 {
			continue
		}

		// Get user settings or use defaults
		settings := userSettings[chatID]
		reminderTime := s.defaultTime
		timezone := s.defaultTimezone.String()

		if settings != nil {
			reminderTime = settings.ReminderTime
			timezone = settings.Timezone
		}

		// Check if it's time to send reminder for this user
		if !s.shouldSendReminderForUser(reminderTime, timezone) {
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
		} else {
			log.Printf("Sent reminder to chat %d at %s %s", chatID, reminderTime, timezone)
		}
	}
}
