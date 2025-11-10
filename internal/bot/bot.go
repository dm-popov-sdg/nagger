package bot

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/dm-popov-sdg/nagger/internal/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot represents the Telegram bot
type Bot struct {
	api     *tgbotapi.BotAPI
	storage *storage.MongoDB
}

// NewBot creates a new Telegram bot instance
func NewBot(token string, storage *storage.MongoDB) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	log.Printf("Authorized on account %s", api.Self.UserName)

	return &Bot{
		api:     api,
		storage: storage,
	}, nil
}

// Start starts the bot
func (b *Bot) Start(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			return nil
		case update := <-updates:
			if update.Message == nil {
				continue
			}

			b.handleMessage(ctx, update.Message)
		}
	}
}

func (b *Bot) handleMessage(ctx context.Context, message *tgbotapi.Message) {
	if !message.IsCommand() {
		return
	}

	switch message.Command() {
	case "start":
		b.handleStart(message)
	case "help":
		b.handleHelp(message)
	case "add":
		b.handleAdd(ctx, message)
	case "list":
		b.handleList(ctx, message)
	case "done":
		b.handleDone(ctx, message)
	case "delete":
		b.handleDelete(ctx, message)
	default:
		b.sendMessage(message.Chat.ID, "Unknown command. Use /help to see available commands.")
	}
}

func (b *Bot) handleStart(message *tgbotapi.Message) {
	text := `Welcome to Nagger Bot! 🤖

I'll help you manage your tasks and remind you about them every day.

Use /help to see available commands.`
	b.sendMessage(message.Chat.ID, text)
}

func (b *Bot) handleHelp(message *tgbotapi.Message) {
	text := `Available commands:

/add <task> - Add a new task
/list - Show all active tasks
/done <task_number> - Mark a task as completed
/delete <task_number> - Delete a task
/help - Show this help message

I'll send you a reminder about your tasks every day at the configured time.`
	b.sendMessage(message.Chat.ID, text)
}

func (b *Bot) handleAdd(ctx context.Context, message *tgbotapi.Message) {
	description := strings.TrimSpace(message.CommandArguments())
	if description == "" {
		b.sendMessage(message.Chat.ID, "Please provide a task description. Usage: /add <task>")
		return
	}

	task := &storage.Task{
		ChatID:      message.Chat.ID,
		UserID:      message.From.ID,
		Description: description,
	}

	if err := b.storage.AddTask(ctx, task); err != nil {
		log.Printf("Error adding task: %v", err)
		b.sendMessage(message.Chat.ID, "Failed to add task. Please try again.")
		return
	}

	b.sendMessage(message.Chat.ID, fmt.Sprintf("✅ Task added: %s", description))
}

func (b *Bot) handleList(ctx context.Context, message *tgbotapi.Message) {
	tasks, err := b.storage.GetTasksByChatID(ctx, message.Chat.ID)
	if err != nil {
		log.Printf("Error getting tasks: %v", err)
		b.sendMessage(message.Chat.ID, "Failed to get tasks. Please try again.")
		return
	}

	if len(tasks) == 0 {
		b.sendMessage(message.Chat.ID, "You have no active tasks. Great job! 🎉")
		return
	}

	var text strings.Builder
	text.WriteString("📋 Your active tasks:\n\n")
	for i, task := range tasks {
		text.WriteString(fmt.Sprintf("%d. %s\n", i+1, task.Description))
	}

	b.sendMessage(message.Chat.ID, text.String())
}

func (b *Bot) handleDone(ctx context.Context, message *tgbotapi.Message) {
	taskNumber, err := b.parseTaskNumber(message.CommandArguments())
	if err != nil {
		b.sendMessage(message.Chat.ID, "Please provide a valid task number. Usage: /done <task_number>")
		return
	}

	tasks, err := b.storage.GetTasksByChatID(ctx, message.Chat.ID)
	if err != nil {
		log.Printf("Error getting tasks: %v", err)
		b.sendMessage(message.Chat.ID, "Failed to get tasks. Please try again.")
		return
	}

	if taskNumber < 1 || taskNumber > len(tasks) {
		b.sendMessage(message.Chat.ID, fmt.Sprintf("Invalid task number. You have %d tasks.", len(tasks)))
		return
	}

	task := tasks[taskNumber-1]
	if err := b.storage.CompleteTask(ctx, task.ID); err != nil {
		log.Printf("Error completing task: %v", err)
		b.sendMessage(message.Chat.ID, "Failed to complete task. Please try again.")
		return
	}

	b.sendMessage(message.Chat.ID, fmt.Sprintf("✅ Task completed: %s", task.Description))
}

func (b *Bot) handleDelete(ctx context.Context, message *tgbotapi.Message) {
	taskNumber, err := b.parseTaskNumber(message.CommandArguments())
	if err != nil {
		b.sendMessage(message.Chat.ID, "Please provide a valid task number. Usage: /delete <task_number>")
		return
	}

	tasks, err := b.storage.GetTasksByChatID(ctx, message.Chat.ID)
	if err != nil {
		log.Printf("Error getting tasks: %v", err)
		b.sendMessage(message.Chat.ID, "Failed to get tasks. Please try again.")
		return
	}

	if taskNumber < 1 || taskNumber > len(tasks) {
		b.sendMessage(message.Chat.ID, fmt.Sprintf("Invalid task number. You have %d tasks.", len(tasks)))
		return
	}

	task := tasks[taskNumber-1]
	if err := b.storage.DeleteTask(ctx, task.ID); err != nil {
		log.Printf("Error deleting task: %v", err)
		b.sendMessage(message.Chat.ID, "Failed to delete task. Please try again.")
		return
	}

	b.sendMessage(message.Chat.ID, fmt.Sprintf("🗑️ Task deleted: %s", task.Description))
}

func (b *Bot) parseTaskNumber(arg string) (int, error) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		return 0, fmt.Errorf("empty argument")
	}
	return strconv.Atoi(arg)
}

func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	sentMsg, err := b.api.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v", err)
		return
	}

	// Track the message for auto-deletion
	ctx := context.Background()
	if err := b.storage.TrackBotMessage(ctx, chatID, sentMsg.MessageID); err != nil {
		log.Printf("Error tracking bot message: %v", err)
	}
}

// SendDailyReminder sends a daily reminder about active tasks
func (b *Bot) SendDailyReminder(ctx context.Context, chatID int64, tasks []string) error {
	if len(tasks) == 0 {
		return nil
	}

	var text strings.Builder
	text.WriteString("🔔 Daily Reminder!\n\n")
	text.WriteString(fmt.Sprintf("You have %d active task(s):\n\n", len(tasks)))
	for i, task := range tasks {
		text.WriteString(fmt.Sprintf("%d. %s\n", i+1, task))
	}
	text.WriteString("\nUse /list to see all tasks or /done <number> to mark them as completed.")

	msg := tgbotapi.NewMessage(chatID, text.String())
	sentMsg, err := b.api.Send(msg)
	if err != nil {
		return err
	}

	// Track the message for auto-deletion
	if err := b.storage.TrackBotMessage(ctx, chatID, sentMsg.MessageID); err != nil {
		log.Printf("Error tracking bot message: %v", err)
	}

	return nil
}

// DeleteMessage deletes a specific message from a chat
func (b *Bot) DeleteMessage(ctx context.Context, chatID int64, messageID int) error {
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
	_, err := b.api.Request(deleteMsg)
	return err
}
