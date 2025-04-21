package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Bot struct {
		Token   string
		Debug   bool
		Timeout int
	}
	Scheduler struct {
		DefaultInterval time.Duration
		RetryAttempts  int
	}
	Notifications struct {
		DefaultChatID int64
		TimeFormat   string
		Templates    map[string]string
	}
	Train struct {
		APIKey    string
		UserAgent string
		BaseURL   string
		Timeout   time.Duration
	}
}

func Load() (*Config, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		// It's okay if .env file doesn't exist
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	cfg := &Config{}

	// Bot configuration
	cfg.Bot.Token = getEnvOrDefault("BOT_TOKEN", "")
	cfg.Bot.Debug = getEnvBoolOrDefault("BOT_DEBUG", false)
	cfg.Bot.Timeout = getEnvIntOrDefault("BOT_TIMEOUT", 60)

	// Scheduler configuration
	cfg.Scheduler.DefaultInterval = time.Hour * 24 * 30 // Default to monthly
	cfg.Scheduler.RetryAttempts = getEnvIntOrDefault("SCHEDULER_RETRY_ATTEMPTS", 3)

	// Notifications configuration
	chatID, err := strconv.ParseInt(getEnvOrDefault("DEFAULT_CHAT_ID", "0"), 10, 64)
	if err != nil {
		return nil, err
	}
	cfg.Notifications.DefaultChatID = chatID
	cfg.Notifications.TimeFormat = getEnvOrDefault("TIME_FORMAT", "Mon Jan 2 15:04:05")
	cfg.Notifications.Templates = map[string]string{
		"monthly": getEnvOrDefault("MONTHLY_TEMPLATE", "need to use sibus"),
	}

	// Train service configuration
	cfg.Train.APIKey = getEnvOrDefault("TRAIN_API_KEY", "")
	cfg.Train.UserAgent = getEnvOrDefault("TRAIN_USER_AGENT", 
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15")
	cfg.Train.BaseURL = getEnvOrDefault("TRAIN_BASE_URL", 
		"https://israelrail.azurefd.net/rjpa-prod/api/v1")
	cfg.Train.Timeout = time.Duration(getEnvIntOrDefault("TRAIN_TIMEOUT", 10)) * time.Second

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
} 