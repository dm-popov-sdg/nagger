package storage

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Task represents a task to be completed
type Task struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	ChatID      int64              `bson:"chat_id"`
	UserID      int64              `bson:"user_id"`
	Description string             `bson:"description"`
	CreatedAt   time.Time          `bson:"created_at"`
	Completed   bool               `bson:"completed"`
}
