package bot

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/dm-popov-sdg/nagger/internal/storage"
	"github.com/dm-popov-sdg/nagger/internal/types"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
			if update.Message != nil {
				b.handleMessage(ctx, update.Message)
			} else if update.CallbackQuery != nil {
				b.handleCallbackQuery(ctx, update.CallbackQuery)
			}
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
	case "setreminder":
		b.handleSetReminder(ctx, message)
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
/done <task_number> - Mark a task as completed for today
/delete <task_number> - Close a task permanently (no more reminders)
/setreminder <HH:MM> [timezone] - Set your daily reminder time (24-hour format)
/help - Show this help message

I'll send you a reminder about your tasks every day at your configured time.

Examples:
/setreminder 09:00 - Set reminder to 9:00 AM UTC
/setreminder 14:30 America/New_York - Set reminder to 2:30 PM EST/EDT`
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
	text.WriteString("📋 Your tasks:\n\n")
	for i, task := range tasks {
		statusEmoji := ""
		switch task.Status {
		case storage.TaskStatusCompletedToday:
			statusEmoji = " ✅"
		case storage.TaskStatusActive, "":
			statusEmoji = ""
		}
		text.WriteString(fmt.Sprintf("%d. %s%s\n", i+1, task.Description, statusEmoji))
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
	if err := b.storage.CloseTask(ctx, task.ID); err != nil {
		log.Printf("Error closing task: %v", err)
		b.sendMessage(message.Chat.ID, "Failed to close task. Please try again.")
		return
	}

	b.sendMessage(message.Chat.ID, fmt.Sprintf("🗑️ Task closed: %s", task.Description))
}

func (b *Bot) handleSetReminder(ctx context.Context, message *tgbotapi.Message) {
	args := strings.Fields(message.CommandArguments())
	if len(args) == 0 {
		b.sendMessage(message.Chat.ID, "Please provide a reminder time. Usage: /setreminder <HH:MM> [timezone]\nExample: /setreminder 09:00 UTC")
		return
	}

	reminderTime := args[0]
	// Validate time format
	if !isValidTimeFormat(reminderTime) {
		b.sendMessage(message.Chat.ID, "Invalid time format. Please use 24-hour format HH:MM (e.g., 09:00, 14:30)")
		return
	}

	// Default timezone is UTC
	timezone := "UTC"
	if len(args) > 1 {
		timezone = args[1]
		// Validate timezone
		if !isValidTimezone(timezone) {
			b.sendMessage(message.Chat.ID, fmt.Sprintf("Invalid timezone: %s. Please use a valid timezone (e.g., UTC, America/New_York)", timezone))
			return
		}
	}

	// Create or update user settings
	settings := &storage.UserSettings{
		ChatID:       message.Chat.ID,
		UserID:       message.From.ID,
		ReminderTime: reminderTime,
		Timezone:     timezone,
	}

	if err := b.storage.SetUserSettings(ctx, settings); err != nil {
		log.Printf("Error setting user settings: %v", err)
		b.sendMessage(message.Chat.ID, "Failed to save reminder settings. Please try again.")
		return
	}

	b.sendMessage(message.Chat.ID, fmt.Sprintf("✅ Reminder time set to %s %s", reminderTime, timezone))
}

func isValidTimeFormat(timeStr string) bool {
	// Check format HH:MM
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return false
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return false
	}

	minute, err := strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return false
	}

	return true
}

func isValidTimezone(tz string) bool {
	// Try to load the timezone
	_, err := time.LoadLocation(tz)
	return err == nil
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
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

// SendDailyReminder sends a daily reminder about active tasks
func (b *Bot) SendDailyReminder(ctx context.Context, chatID int64, tasks []string) error {
	if len(tasks) == 0 {
		return nil
	}

	var text strings.Builder
	text.WriteString("🔔 Daily Reminder!\n\n")
	text.WriteString(fmt.Sprintf("You have %d active task(s):", len(tasks)))

	msg := tgbotapi.NewMessage(chatID, text.String())
	_, err := b.api.Send(msg)
	return err
}

// SendDailyReminderWithTasks sends a daily reminder with inline keyboard for task completion
func (b *Bot) SendDailyReminderWithTasks(ctx context.Context, chatID int64, tasks []types.TaskWithID) error {
	if len(tasks) == 0 {
		return nil
	}

	var text strings.Builder
	text.WriteString("🔔 Daily Reminder!\n\n")
	text.WriteString(fmt.Sprintf("You have %d active task(s). Click on a task to mark it as done:", len(tasks)))

	// Create inline keyboard with tasks
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, task := range tasks {
		statusEmoji := "⬜"
		if task.GetStatus() == string(storage.TaskStatusCompletedToday) {
			statusEmoji = "✅"
		}
		buttonText := fmt.Sprintf("%s %s", statusEmoji, task.GetDescription())
		buttonData := fmt.Sprintf("complete_%s", task.GetID())
		button := tgbotapi.NewInlineKeyboardButtonData(buttonText, buttonData)
		row := tgbotapi.NewInlineKeyboardRow(button)
		rows = append(rows, row)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ReplyMarkup = keyboard
	_, err := b.api.Send(msg)
	return err
}

func (b *Bot) handleCallbackQuery(ctx context.Context, query *tgbotapi.CallbackQuery) {
	// Acknowledge the callback query
	callback := tgbotapi.NewCallback(query.ID, "")
	if _, err := b.api.Request(callback); err != nil {
		log.Printf("Error acknowledging callback: %v", err)
	}

	// Check if this is a task completion callback
	if strings.HasPrefix(query.Data, "complete_") {
		taskIDHex := strings.TrimPrefix(query.Data, "complete_")
		taskID, err := primitive.ObjectIDFromHex(taskIDHex)
		if err != nil {
			log.Printf("Invalid task ID in callback: %v", err)
			return
		}

		// Get the task to check its current status
		tasks, err := b.storage.GetTasksByChatID(ctx, query.Message.Chat.ID)
		if err != nil {
			log.Printf("Error getting tasks: %v", err)
			return
		}

		var task *storage.Task
		for i := range tasks {
			if tasks[i].ID == taskID {
				task = &tasks[i]
				break
			}
		}

		if task == nil {
			log.Printf("Task not found: %s", taskIDHex)
			return
		}

		// Toggle task status
		if task.Status == storage.TaskStatusCompletedToday {
			// Reactivate the task
			err = b.storage.ReactivateTask(ctx, taskID)
			if err != nil {
				log.Printf("Error reactivating task: %v", err)
				return
			}
		} else {
			// Complete the task
			err = b.storage.CompleteTask(ctx, taskID)
			if err != nil {
				log.Printf("Error completing task: %v", err)
				return
			}
		}

		// Get updated tasks and rebuild the keyboard
		updatedTasks, err := b.storage.GetTasksByChatID(ctx, query.Message.Chat.ID)
		if err != nil {
			log.Printf("Error getting updated tasks: %v", err)
			return
		}

		// Rebuild inline keyboard with updated status
		var rows [][]tgbotapi.InlineKeyboardButton
		for _, t := range updatedTasks {
			statusEmoji := "⬜"
			if t.Status == storage.TaskStatusCompletedToday {
				statusEmoji = "✅"
			}
			buttonText := fmt.Sprintf("%s %s", statusEmoji, t.Description)
			buttonData := fmt.Sprintf("complete_%s", t.ID.Hex())
			button := tgbotapi.NewInlineKeyboardButtonData(buttonText, buttonData)
			row := tgbotapi.NewInlineKeyboardRow(button)
			rows = append(rows, row)
		}

		keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
		edit := tgbotapi.NewEditMessageReplyMarkup(
			query.Message.Chat.ID,
			query.Message.MessageID,
			keyboard,
		)
		if _, err := b.api.Send(edit); err != nil {
			log.Printf("Error updating message: %v", err)
		}
	}
}
