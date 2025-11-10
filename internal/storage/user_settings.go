package storage

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserSettings represents user-specific settings
type UserSettings struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	ChatID       int64              `bson:"chat_id"`
	UserID       int64              `bson:"user_id"`
	ReminderTime string             `bson:"reminder_time"` // Format: "HH:MM" (24-hour format)
	Timezone     string             `bson:"timezone"`      // e.g., "UTC", "America/New_York"
	CreatedAt    time.Time          `bson:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at"`
}
