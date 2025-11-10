# Nagger - Telegram Task Reminder Bot

A Telegram bot that helps you manage tasks and sends daily reminders. Tasks are stored in MongoDB, and you'll receive automated reminders about pending tasks at the start of each day.

## Features

- ✅ Add, list, complete, and delete tasks via Telegram
- 📅 Daily reminders about active tasks
- 💾 Persistent storage using MongoDB
- 🐳 Docker support for easy deployment
- ⚙️ Per-user configurable reminder time and timezone
- 🌍 Support for any timezone worldwide

## Commands

- `/start` - Start the bot and see welcome message
- `/help` - Show available commands
- `/add <task>` - Add a new task
- `/list` - Show all tasks (active and completed today)
- `/done <task_number>` - Mark a task as completed for today
- `/delete <task_number>` - Close a task permanently (no more reminders)
- `/setreminder <HH:MM> [timezone]` - Set your personal reminder time (24-hour format)

### Setting Your Reminder Time

Each user can set their own reminder time and timezone using the `/setreminder` command:

```
/setreminder 09:00          # Set to 9:00 AM UTC (default timezone)
/setreminder 14:30 UTC      # Set to 2:30 PM UTC
/setreminder 08:00 America/New_York  # Set to 8:00 AM Eastern Time
/setreminder 22:00 Europe/London     # Set to 10:00 PM London Time
```

If you don't set a reminder time, the bot will use the default time specified in the environment variables.

## Configuration

The bot is configured using environment variables:

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `TELEGRAM_BOT_TOKEN` | Your Telegram bot token from @BotFather | - | Yes |
| `MONGO_URI` | MongoDB connection string | - | Yes |
| `MONGO_DB` | MongoDB database name | `nagger` | No |
| `REMINDER_TIME` | Default reminder time for users who haven't set their own (24-hour format HH:MM) | `09:00` | No |
| `REMINDER_TIMEZONE` | Default timezone for users who haven't set their own (e.g., UTC, America/New_York) | `UTC` | No |

**Note:** Users can override the default reminder time and timezone by using the `/setreminder` command.

## Getting Started

### Prerequisites

- Go 1.21 or higher (for local development)
- Docker and Docker Compose (for containerized deployment)
- A Telegram bot token (get one from [@BotFather](https://t.me/botfather))
- MongoDB instance or Docker

### Local Development

1. Clone the repository:
```bash
git clone https://github.com/dm-popov-sdg/nagger.git
cd nagger
```

2. Create a `.env` file from the example:
```bash
cp .env.example .env
```

3. Edit `.env` and add your Telegram bot token and MongoDB connection string.

4. Install dependencies:
```bash
go mod download
```

5. Run the bot:
```bash
go run cmd/bot/main.go
```

### Docker Deployment

1. Clone the repository:
```bash
git clone https://github.com/dm-popov-sdg/nagger.git
cd nagger
```

2. Create a `.env` file:
```bash
cp .env.example .env
```

3. Edit `.env` and set your `TELEGRAM_BOT_TOKEN`.

4. Start the services using Docker Compose:
```bash
docker-compose up -d
```

This will start both the bot and a MongoDB instance.

To view logs:
```bash
docker-compose logs -f bot
```

To stop the services:
```bash
docker-compose down
```

### Building Docker Image Manually

To build the Docker image:
```bash
docker build -t nagger-bot .
```

To run the container:
```bash
docker run -d \
  --name nagger-bot \
  -e TELEGRAM_BOT_TOKEN=your_token \
  -e MONGO_URI=mongodb://your_mongo_host:27017/ \
  nagger-bot
```

## How It Works

1. **Task Management**: Users interact with the bot via Telegram commands to manage their tasks.
2. **Task States**: Tasks have three states:
   - **Active**: New tasks that need to be done
   - **Completed Today**: Tasks marked as done with `/done` - they still appear in reminders for recurring daily tasks
   - **Closed**: Tasks closed with `/delete` - they no longer appear in reminders
3. **Storage**: All tasks and user settings are stored in MongoDB with information about the chat, user, description, status, and personal reminder preferences.
4. **Daily Reminders**: A scheduler runs in the background and sends reminders to each user at their configured time (or the default time if not set). Each user can set their own reminder time and timezone using `/setreminder`.

## MongoDB Connection String Format

The `MONGO_URI` should be in the standard MongoDB connection string format:

```
mongodb://[username:password@]host[:port][/[database][?options]]
```

Examples:
- Local: `mongodb://localhost:27017/`
- With auth: `mongodb://user:pass@localhost:27017/`
- MongoDB Atlas: `mongodb+srv://user:pass@cluster.mongodb.net/`

## Development

### Project Structure

```
nagger/
├── cmd/
│   └── bot/           # Main application entry point
│       └── main.go
├── internal/
│   ├── bot/           # Telegram bot implementation
│   ├── config/        # Configuration management
│   ├── scheduler/     # Daily reminder scheduler
│   └── storage/       # MongoDB storage layer
├── Dockerfile         # Docker image definition
├── docker-compose.yml # Docker Compose configuration
└── README.md
```

## License

See [LICENSE](LICENSE) file for details.
