package train

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"rail-go/internal/config"
)

// AzureEndpoint implements the Endpoint interface for Azure API
type AzureEndpoint struct {
	url    string
	config *config.Config
}

// NewAzureEndpoint creates a new Azure endpoint
func NewAzureEndpoint(url string, cfg *config.Config) *AzureEndpoint {
	return &AzureEndpoint{
		url:    url,
		config: cfg,
	}
}

func (e *AzureEndpoint) BuildRequest(ctx context.Context, from, to string) (*http.Request, error) {
	params := url.Values{}
	params.Add("fromStation", from)
	params.Add("toStation", to)
	params.Add("date", time.Now().Format("2006-01-02"))
	params.Add("hour", time.Now().Format("15:04:05"))
	params.Add("scheduleType", "2")
	params.Add("systemType", "1")
	params.Add("languageId", "Hebrew")

	req, err := http.NewRequestWithContext(ctx, "GET", e.url, nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = params.Encode()
	return req, nil
}

func (e *AzureEndpoint) ParseResponse(data []byte) (string, error) {
	var schedule ScheduleResponse
	if err := json.Unmarshal(data, &schedule); err != nil {
		return "", fmt.Errorf("failed to decode azure response: %w", err)
	}
	return formatSchedule(schedule), nil
}

func (e *AzureEndpoint) RequiresAPIKey() bool {
	return true
}

func (e *AzureEndpoint) GetURL() string {
	return e.url
}

func (e *AzureEndpoint) GetFormat() string {
	return "azure"
}

// MobileEndpoint implements the Endpoint interface for Mobile API
type MobileEndpoint struct {
	url    string
	config *config.Config
}

// NewMobileEndpoint creates a new Mobile endpoint
func NewMobileEndpoint(url string, cfg *config.Config) *MobileEndpoint {
	return &MobileEndpoint{
		url:    url,
		config: cfg,
	}
}

func (e *MobileEndpoint) BuildRequest(ctx context.Context, from, to string) (*http.Request, error) {
	params := url.Values{}
	params.Add("origin", from)
	params.Add("destination", to)
	params.Add("date", time.Now().Format("02/01/2006 15:04"))
	params.Add("hours", "12")

	req, err := http.NewRequestWithContext(ctx, "GET", e.url, nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = params.Encode()
	return req, nil
}

func (e *MobileEndpoint) ParseResponse(data []byte) (string, error) {
	// Try mobile format first
	var mobileSchedule MobileScheduleResponse
	if err := json.Unmarshal(data, &mobileSchedule); err == nil {
		return formatMobileSchedule(mobileSchedule), nil
	}
	
	// Fallback to azure format
	var schedule ScheduleResponse
	if err := json.Unmarshal(data, &schedule); err != nil {
		return "", fmt.Errorf("failed to decode mobile response: %w", err)
	}
	return formatSchedule(schedule), nil
}

func (e *MobileEndpoint) RequiresAPIKey() bool {
	return false
}

func (e *MobileEndpoint) GetURL() string {
	return e.url
}

func (e *MobileEndpoint) GetFormat() string {
	return "mobile"
}