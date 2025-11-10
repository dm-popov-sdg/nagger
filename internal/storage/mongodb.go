package storage

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB implements task storage using MongoDB
type MongoDB struct {
	client             *mongo.Client
	collection         *mongo.Collection
	settingsCollection *mongo.Collection
}

// NewMongoDB creates a new MongoDB storage instance
func NewMongoDB(ctx context.Context, uri, dbName string) (*MongoDB, error) {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	collection := client.Database(dbName).Collection("tasks")
	settingsCollection := client.Database(dbName).Collection("user_settings")

	return &MongoDB{
		client:             client,
		collection:         collection,
		settingsCollection: settingsCollection,
	}, nil
}

// Close closes the MongoDB connection
func (m *MongoDB) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

// AddTask adds a new task to the storage
func (m *MongoDB) AddTask(ctx context.Context, task *Task) error {
	task.CreatedAt = time.Now()
	task.Completed = false
	task.Status = TaskStatusActive

	result, err := m.collection.InsertOne(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to insert task: %w", err)
	}

	task.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetTasksByChatID retrieves all active tasks for a specific chat
func (m *MongoDB) GetTasksByChatID(ctx context.Context, chatID int64) ([]Task, error) {
	// Get tasks that are not closed (includes active and completed_today)
	filter := bson.M{
		"chat_id": chatID,
		"$or": []bson.M{
			{"status": bson.M{"$ne": TaskStatusClosed}},
			{"status": bson.M{"$exists": false}}, // For backward compatibility with old documents
		},
	}

	cursor, err := m.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks: %w", err)
	}
	defer cursor.Close(ctx)

	var tasks []Task
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, fmt.Errorf("failed to decode tasks: %w", err)
	}

	return tasks, nil
}

// GetAllActiveTasks retrieves all active tasks across all chats
// This excludes only closed tasks - includes both active and completed_today tasks
func (m *MongoDB) GetAllActiveTasks(ctx context.Context) (map[int64][]Task, error) {
	// Get tasks that are not closed (includes active and completed_today)
	filter := bson.M{
		"$or": []bson.M{
			{"status": bson.M{"$ne": TaskStatusClosed}},
			{"status": bson.M{"$exists": false}}, // For backward compatibility with old documents
		},
	}

	cursor, err := m.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks: %w", err)
	}
	defer cursor.Close(ctx)

	var tasks []Task
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, fmt.Errorf("failed to decode tasks: %w", err)
	}

	// Group tasks by chat ID
	tasksByChat := make(map[int64][]Task)
	for _, task := range tasks {
		tasksByChat[task.ChatID] = append(tasksByChat[task.ChatID], task)
	}

	return tasksByChat, nil
}

// CompleteTask marks a task as completed today
func (m *MongoDB) CompleteTask(ctx context.Context, taskID primitive.ObjectID) error {
	filter := bson.M{"_id": taskID}
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"completed":    true,
			"status":       TaskStatusCompletedToday,
			"completed_at": now,
		},
	}

	result, err := m.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// ReactivateTask marks a completed task as active again
func (m *MongoDB) ReactivateTask(ctx context.Context, taskID primitive.ObjectID) error {
	filter := bson.M{"_id": taskID}
	update := bson.M{
		"$set": bson.M{
			"completed": false,
			"status":    TaskStatusActive,
		},
		"$unset": bson.M{
			"completed_at": "",
		},
	}

	result, err := m.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// CloseTask marks a task as permanently closed (no more reminders)
func (m *MongoDB) CloseTask(ctx context.Context, taskID primitive.ObjectID) error {
	filter := bson.M{"_id": taskID}
	update := bson.M{
		"$set": bson.M{
			"completed": true,
			"status":    TaskStatusClosed,
		},
	}

	result, err := m.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// DeleteTask removes a task from storage
func (m *MongoDB) DeleteTask(ctx context.Context, taskID primitive.ObjectID) error {
	filter := bson.M{"_id": taskID}

	result, err := m.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// GetUserSettings retrieves user settings for a specific chat
func (m *MongoDB) GetUserSettings(ctx context.Context, chatID int64) (*UserSettings, error) {
	filter := bson.M{"chat_id": chatID}

	var settings UserSettings
	err := m.settingsCollection.FindOne(ctx, filter).Decode(&settings)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // No settings found, will use defaults
		}
		return nil, fmt.Errorf("failed to find user settings: %w", err)
	}

	return &settings, nil
}

// SetUserSettings creates or updates user settings for a specific chat
func (m *MongoDB) SetUserSettings(ctx context.Context, settings *UserSettings) error {
	settings.UpdatedAt = time.Now()

	filter := bson.M{"chat_id": settings.ChatID}
	update := bson.M{
		"$set": bson.M{
			"user_id":       settings.UserID,
			"reminder_time": settings.ReminderTime,
			"timezone":      settings.Timezone,
			"updated_at":    settings.UpdatedAt,
		},
		"$setOnInsert": bson.M{
			"created_at": time.Now(),
		},
	}

	opts := options.Update().SetUpsert(true)
	result, err := m.settingsCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to update user settings: %w", err)
	}

	// If this was an insert, set the ID
	if result.UpsertedID != nil {
		settings.ID = result.UpsertedID.(primitive.ObjectID)
	}

	return nil
}

// GetAllUserSettings retrieves all user settings
func (m *MongoDB) GetAllUserSettings(ctx context.Context) (map[int64]*UserSettings, error) {
	cursor, err := m.settingsCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to find user settings: %w", err)
	}
	defer cursor.Close(ctx)

	var settingsList []UserSettings
	if err := cursor.All(ctx, &settingsList); err != nil {
		return nil, fmt.Errorf("failed to decode user settings: %w", err)
	}

	// Group settings by chat ID
	settingsByChat := make(map[int64]*UserSettings)
	for i := range settingsList {
		settingsByChat[settingsList[i].ChatID] = &settingsList[i]
	}

	return settingsByChat, nil
}
