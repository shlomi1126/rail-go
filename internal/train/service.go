package train

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"rail-go/internal/config"
	"rail-go/internal/logger"
)

type Service struct {
	config *config.Config
	logger *logger.Logger
	cache  *sync.Map
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
		cache:  &sync.Map{},
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *Service) GetSchedule(ctx context.Context, from, to string) ([]string, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("%s-%s", from, to)
	if cached, ok := s.cache.Load(cacheKey); ok {
		return cached.([]string), nil
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

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var schedule ScheduleResponse
	if err := json.NewDecoder(resp.Body).Decode(&schedule); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Format response and split into chunks
	result := s.formatSchedule(schedule)
	chunks := splitMessage(result)

	// Cache result
	s.cache.Store(cacheKey, chunks)

	return chunks, nil
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
