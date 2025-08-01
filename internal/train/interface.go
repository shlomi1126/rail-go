package train

import (
	"context"
	"net/http"
)

// HTTPClient interface for making HTTP requests
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Endpoint interface for different API endpoints
type Endpoint interface {
	BuildRequest(ctx context.Context, from, to string) (*http.Request, error)
	ParseResponse(data []byte) (string, error)
	RequiresAPIKey() bool
	GetURL() string
	GetFormat() string
}

// ScheduleProvider interface for getting train schedules
type ScheduleProvider interface {
	GetSchedule(ctx context.Context, from, to string) ([]string, error)
}