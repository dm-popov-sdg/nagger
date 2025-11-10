package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dm-popov-sdg/nagger/internal/bot"
	"github.com/dm-popov-sdg/nagger/internal/config"
	"github.com/dm-popov-sdg/nagger/internal/scheduler"
	"github.com/dm-popov-sdg/nagger/internal/storage"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create context that listens for termination signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize MongoDB storage
	mongodb, err := storage.NewMongoDB(ctx, cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := mongodb.Close(ctx); err != nil {
			log.Printf("Error closing MongoDB connection: %v", err)
		}
	}()

	log.Println("Successfully connected to MongoDB")

	// Create Telegram bot
	telegramBot, err := bot.NewBot(cfg.TelegramToken, mongodb)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Create scheduler
	sched, err := scheduler.NewScheduler(
		&storageAdapter{mongodb},
		telegramBot,
		cfg.ReminderTime,
		cfg.ReminderTimezone,
	)
	if err != nil {
		log.Fatalf("Failed to create scheduler: %v", err)
	}

	// Start scheduler
	sched.Start(ctx)
	defer sched.Stop()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start bot in a goroutine
	errChan := make(chan error, 1)
	go func() {
		log.Println("Starting bot...")
		if err := telegramBot.Start(ctx); err != nil {
			errChan <- err
		}
	}()

	// Wait for termination signal or error
	select {
	case <-sigChan:
		log.Println("Received termination signal, shutting down...")
		cancel()
	case err := <-errChan:
		log.Printf("Bot error: %v", err)
		cancel()
	}

	log.Println("Bot stopped")
}

// storageAdapter adapts storage.MongoDB to scheduler.TaskGetter interface
type storageAdapter struct {
	*storage.MongoDB
}

func (s *storageAdapter) GetAllActiveTasks(ctx context.Context) (map[int64][]scheduler.Task, error) {
	tasks, err := s.MongoDB.GetAllActiveTasks(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to scheduler.Task interface
	result := make(map[int64][]scheduler.Task)
	for chatID, chatTasks := range tasks {
		schedulerTasks := make([]scheduler.Task, len(chatTasks))
		for i, task := range chatTasks {
			schedulerTasks[i] = task
		}
		result[chatID] = schedulerTasks
	}

	return result, nil
}
