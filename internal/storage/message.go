package storage

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BotMessage represents a message sent by the bot that should be auto-deleted
type BotMessage struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	ChatID    int64              `bson:"chat_id"`
	MessageID int                `bson:"message_id"`
	SentAt    time.Time          `bson:"sent_at"`
}
