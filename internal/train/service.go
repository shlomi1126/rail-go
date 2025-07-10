package train

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"rail-go/internal/cache"
	"rail-go/internal/config"
	"rail-go/internal/logger"
)

type Service struct {
	config *config.Config
	logger *logger.Logger
	cache  *cache.Cache
	client *http.Client
}

type ScheduleResponse struct {
	Result struct {
		Travels []struct {
			Trains []TrainRoutePart `json:"trains"`
		} `json:"travels"`
	} `json:"result"`
}

type TrainRoutePart struct {
	OriginStation      int    `json:"orignStation"`
	DestinationStation int    `json:"destinationStation"`
	ArrivalTime        string `json:"arrivalTime"`
	DepartureTime      string `json:"departureTime"`
	OriginPlatform     int    `json:"originPlatform"`
	DestPlatform       int    `json:"destPlatform"`
}

func NewService(cfg *config.Config, log *logger.Logger) *Service {
	return &Service{
		config: cfg,
		logger: log,
		cache:  cache.NewCache(),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *Service) GetSchedule(ctx context.Context, from, to string) ([]string, error) {
	return s.getScheduleWithRetry(ctx, from, to, 3)
}

func (s *Service) getScheduleWithRetry(ctx context.Context, from, to string, maxRetries int) ([]string, error) {
	// Check cache first
	if cached := s.cache.Get("", from, to); cached != nil {
		if chunks, ok := cached.([]string); ok {
			return chunks, nil
		}
	}

	// Build request
	params := url.Values{}
	params.Add("fromStation", from)
	params.Add("toStation", to)
	params.Add("date", time.Now().Format("2006-01-02"))
	params.Add("hour", time.Now().Format("15:04:05"))
	params.Add("scheduleType", "2")
	params.Add("systemType", "1")
	params.Add("languageId", "Hebrew")

	// Make request
	req, err := http.NewRequestWithContext(ctx, "GET",
		"https://israelrail.azurefd.net/rjpa-prod/api/v1/timetable/searchTrainLuzForDateTime", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.URL.RawQuery = params.Encode()
	req.Header.Set("User-Agent", s.config.Train.UserAgent)
	req.Header.Set("ocp-apim-subscription-key", s.config.Train.APIKey)

	s.logger.Info("Making API request",
		"from_station", from,
		"to_station", to,
		"url", req.URL.String())
	
	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Error("HTTP request failed",
			"error", err,
			"from_station", from,
			"to_station", to)
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Log detailed error information
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error("API request failed",
			"status_code", resp.StatusCode,
			"response_body", string(body),
			"from_station", from,
			"to_station", to,
			"url", req.URL.String())
		if maxRetries > 0 && s.shouldRetry(resp.StatusCode) {
			s.logger.Info("Retrying API request",
				"retries_left", maxRetries-1,
				"status_code", resp.StatusCode)
			time.Sleep(2 * time.Second) // Wait before retry
			return s.getScheduleWithRetry(ctx, from, to, maxRetries-1)
		}
		return nil, fmt.Errorf("API request failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var schedule ScheduleResponse
	if err := json.NewDecoder(resp.Body).Decode(&schedule); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Format response and split into chunks
	result := s.formatSchedule(schedule)
	chunks := splitMessage(result)

	// Cache result
	s.cache.Set("", from, to, chunks)

	return chunks, nil
}

func (s *Service) shouldRetry(statusCode int) bool {
	// Retry on server errors (5xx) and rate limiting (429)
	return statusCode >= 500 || statusCode == 429
}

func (s *Service) formatSchedule(schedule ScheduleResponse) string {
	var result string
	count := 0
	for i, travel := range schedule.Result.Travels {
		if count >= 5 {
			break
		}
		result += fmt.Sprintf("🚆 %d:\n", i+1)
		for j, train := range travel.Trains {
			if count >= 5 {
				break
			}
			result += fmt.Sprintf("  🚂 %d:\n", j+1)
			result += fmt.Sprintf("    עליה: %s (רציף %d)\n",
				s.GetStationName(fmt.Sprintf("%d", train.OriginStation)), train.OriginPlatform)
			result += fmt.Sprintf("    זמן יציאת הרכבת: %s\n",
				s.formatTime(train.DepartureTime))
			result += fmt.Sprintf("    אל: %s (רציף %d)\n",
				s.GetStationName(fmt.Sprintf("%d", train.DestinationStation)), train.DestPlatform)
			result += fmt.Sprintf("    זמן הגעה: %s\n",
				s.formatTime(train.ArrivalTime))
			count++
		}
		result += "\n"
	}
	return result
}

func (s *Service) GetStationName(stationID string) string {
	if station, ok := STATIONS[stationID]; ok {
		return station["Heb"]
	}
	return stationID
}

func (s *Service) GetStationSuggestions(query string) map[string]string {
	suggestions := make(map[string]string)
	query = strings.ToLower(query)

	for id, station := range STATIONS {
		// Check Hebrew name
		if strings.Contains(strings.ToLower(station["Heb"]), query) {
			suggestions[station["Heb"]] = id
		}
		// Check English name
		if strings.Contains(strings.ToLower(station["Eng"]), query) {
			suggestions[station["Heb"]] = id
		}
	}

	return suggestions
}

func (s *Service) formatTime(timeStr string) string {
	t, err := time.Parse("2006-01-02T15:04:05", timeStr)
	if err != nil {
		return timeStr
	}
	return t.Format("15:04:05")
}

func splitMessage(message string) []string {
	const maxLength = 4000
	var chunks []string
	currentChunk := ""
	lines := strings.Split(message, "\n")

	for _, line := range lines {
		if len(currentChunk)+len(line)+1 > maxLength {
			chunks = append(chunks, currentChunk)
			currentChunk = line + "\n"
		} else {
			currentChunk += line + "\n"
		}
	}

	if currentChunk != "" {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}
