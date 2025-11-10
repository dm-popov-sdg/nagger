package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	TelegramToken    string
	MongoURI         string
	MongoDB          string
	ReminderTime     string // Format: "HH:MM" (24-hour format)
	ReminderTimezone string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		TelegramToken:    os.Getenv("TELEGRAM_BOT_TOKEN"),
		MongoURI:         os.Getenv("MONGO_URI"),
		MongoDB:          getEnvOrDefault("MONGO_DB", "nagger"),
		ReminderTime:     getEnvOrDefault("REMINDER_TIME", "09:00"),
		ReminderTimezone: getEnvOrDefault("REMINDER_TIMEZONE", "UTC"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if required configuration values are set
func (c *Config) Validate() error {
	if c.TelegramToken == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}
	if c.MongoURI == "" {
		return fmt.Errorf("MONGO_URI is required")
	}
	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
