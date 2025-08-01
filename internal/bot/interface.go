package bot

import "context"

// TrainService interface for train schedule operations
type TrainService interface {
	GetSchedule(ctx context.Context, from, to string) ([]string, error)
	GetStationName(stationID string) string
	GetStationSuggestions(query string) map[string]string
}