package storage

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	// TaskStatusActive means the task is active and needs to be done
	TaskStatusActive TaskStatus = "active"
	// TaskStatusCompletedToday means the task was completed today but can still be reminded about
	TaskStatusCompletedToday TaskStatus = "completed_today"
	// TaskStatusClosed means the task is permanently closed and should not be reminded about
	TaskStatusClosed TaskStatus = "closed"
)

// Task represents a task to be completed
type Task struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	ChatID      int64              `bson:"chat_id"`
	UserID      int64              `bson:"user_id"`
	Description string             `bson:"description"`
	CreatedAt   time.Time          `bson:"created_at"`
	Completed   bool               `bson:"completed"` // Deprecated: kept for backward compatibility
	Status      TaskStatus         `bson:"status"`
	CompletedAt *time.Time         `bson:"completed_at,omitempty"` // When the task was completed
}
