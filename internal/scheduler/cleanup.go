package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/dm-popov-sdg/nagger/internal/storage"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MessageDeleter defines the interface for deleting messages
type MessageDeleter interface {
	DeleteMessage(ctx context.Context, chatID int64, messageID int) error
}

// MessageStorage defines the interface for message storage operations
type MessageStorage interface {
	GetMessagesOlderThan(ctx context.Context, olderThan time.Time) ([]storage.BotMessage, error)
	DeleteBotMessage(ctx context.Context, messageID primitive.ObjectID) error
}

// CleanupScheduler handles automatic deletion of old bot messages
type CleanupScheduler struct {
	storage       MessageStorage
	bot           MessageDeleter
	cleanupPeriod time.Duration
	messageAge    time.Duration
	stopChan      chan struct{}
}

// NewCleanupScheduler creates a new cleanup scheduler
func NewCleanupScheduler(storage MessageStorage, bot MessageDeleter, messageAge time.Duration) *CleanupScheduler {
	return &CleanupScheduler{
		storage:       storage,
		bot:           bot,
		cleanupPeriod: 1 * time.Hour, // Run cleanup every hour
		messageAge:    messageAge,
		stopChan:      make(chan struct{}),
	}
}

// Start begins the cleanup scheduler
func (c *CleanupScheduler) Start(ctx context.Context) {
	go c.run(ctx)
}

// Stop stops the cleanup scheduler
func (c *CleanupScheduler) Stop() {
	close(c.stopChan)
}

func (c *CleanupScheduler) run(ctx context.Context) {
	ticker := time.NewTicker(c.cleanupPeriod)
	defer ticker.Stop()

	log.Printf("Cleanup scheduler started. Will delete messages older than %v", c.messageAge)

	// Run cleanup immediately on start
	c.cleanupOldMessages(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopChan:
			return
		case <-ticker.C:
			c.cleanupOldMessages(ctx)
		}
	}
}

func (c *CleanupScheduler) cleanupOldMessages(ctx context.Context) {
	log.Println("Running message cleanup...")

	cutoffTime := time.Now().Add(-c.messageAge)
	messages, err := c.storage.GetMessagesOlderThan(ctx, cutoffTime)
	if err != nil {
		log.Printf("Error getting old messages: %v", err)
		return
	}

	if len(messages) == 0 {
		log.Println("No old messages to clean up")
		return
	}

	log.Printf("Found %d messages to delete", len(messages))

	deletedCount := 0
	for _, msg := range messages {
		// Try to delete the message from Telegram
		if err := c.bot.DeleteMessage(ctx, msg.ChatID, msg.MessageID); err != nil {
			log.Printf("Error deleting message %d in chat %d: %v", msg.MessageID, msg.ChatID, err)
			// Continue anyway - the message might have been already deleted
		} else {
			deletedCount++
		}

		// Remove the message record from storage regardless of deletion success
		if err := c.storage.DeleteBotMessage(ctx, msg.ID); err != nil {
			log.Printf("Error removing message record: %v", err)
		}
	}

	log.Printf("Message cleanup completed. Successfully deleted %d messages", deletedCount)
}
