# Nagger - Telegram Task Reminder Bot

A Telegram bot that helps you manage tasks and sends daily reminders. Tasks are stored in MongoDB, and you'll receive automated reminders about pending tasks at the start of each day.

## Features

- ✅ Add, list, complete, and delete tasks via Telegram
- 📅 Daily reminders about active tasks
- 💾 Persistent storage using MongoDB
- 🐳 Docker support for easy deployment
- ⚙️ Configurable reminder time and timezone

## Commands

- `/start` - Start the bot and see welcome message
- `/help` - Show available commands
- `/add <task>` - Add a new task
- `/list` - Show all active tasks
- `/done <task_number>` - Mark a task as completed
- `/delete <task_number>` - Delete a task

## Configuration

The bot is configured using environment variables:

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `TELEGRAM_BOT_TOKEN` | Your Telegram bot token from @BotFather | - | Yes |
| `MONGO_URI` | MongoDB connection string | - | Yes |
| `MONGO_DB` | MongoDB database name | `nagger` | No |
| `REMINDER_TIME` | Daily reminder time (24-hour format HH:MM) | `09:00` | No |
| `REMINDER_TIMEZONE` | Timezone for reminders (e.g., UTC, America/New_York) | `UTC` | No |

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
2. **Storage**: All tasks are stored in MongoDB with information about the chat, user, description, and completion status.
3. **Daily Reminders**: A scheduler runs in the background and sends reminders to all users with active tasks at the configured time each day.

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
