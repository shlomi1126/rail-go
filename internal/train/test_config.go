package train

import (
	"time"

	"rail-go/internal/config"
)

// NewTestConfig creates a mock configuration for testing
func NewTestConfig() *config.Config {
	cfg := &config.Config{}

	// Bot configuration
	cfg.Bot.Token = "test_token"
	cfg.Bot.Debug = true
	cfg.Bot.Timeout = 60

	// Scheduler configuration
	cfg.Scheduler.DefaultInterval = time.Hour * 24 * 30 // Default to monthly
	cfg.Scheduler.RetryAttempts = 3

	// Notifications configuration
	cfg.Notifications.DefaultChatID = 0
	cfg.Notifications.TimeFormat = "Mon Jan 2 15:04:05"
	cfg.Notifications.Templates = map[string]string{
		"monthly": "need to use sibus",
	}

	// Train service configuration
	cfg.Train.APIKey = "test_api_key"
	cfg.Train.UserAgent = "test_user_agent"
	cfg.Train.BaseURL = "https://test.api.rail.co.il"
	cfg.Train.Timeout = 30 * time.Second

	return cfg
}
