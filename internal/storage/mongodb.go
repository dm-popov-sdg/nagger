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
	client     *mongo.Client
	collection *mongo.Collection
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

	return &MongoDB{
		client:     client,
		collection: collection,
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

	result, err := m.collection.InsertOne(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to insert task: %w", err)
	}

	task.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetTasksByChatID retrieves all active tasks for a specific chat
func (m *MongoDB) GetTasksByChatID(ctx context.Context, chatID int64) ([]Task, error) {
	filter := bson.M{
		"chat_id":   chatID,
		"completed": false,
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
func (m *MongoDB) GetAllActiveTasks(ctx context.Context) (map[int64][]Task, error) {
	filter := bson.M{"completed": false}

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

// CompleteTask marks a task as completed
func (m *MongoDB) CompleteTask(ctx context.Context, taskID primitive.ObjectID) error {
	filter := bson.M{"_id": taskID}
	update := bson.M{"$set": bson.M{"completed": true}}

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
