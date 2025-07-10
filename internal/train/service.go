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

	"go.uber.org/zap"
	"rail-go/internal/cache"
	"rail-go/internal/config"
	"rail-go/internal/logger"
)

type APIEndpoint struct {
	URL        string
	Format     string // "azure" or "mobile"
	RequiresKey bool
}

type Service struct {
	config    *config.Config
	logger    *logger.Logger
	cache     *cache.Cache
	client    *http.Client
	endpoints []APIEndpoint
}

type ScheduleResponse struct {
	Result struct {
		Travels []struct {
			Trains []TrainRoutePart `json:"trains"`
		} `json:"travels"`
	} `json:"result"`
}

// Mobile API response format (alternative structure)
type MobileScheduleResponse struct {
	Data []struct {
		OriginStationId      int    `json:"OriginStationId"`
		DestinationStationId int    `json:"DestinationStationId"`
		DepartureTime        string `json:"DepartureTime"`
		ArrivalTime          string `json:"ArrivalTime"`
		OriginPlatform       int    `json:"OriginPlatform"`
		DestinationPlatform  int    `json:"DestinationPlatform"`
	} `json:"data"`
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
	endpoints := []APIEndpoint{
		// Primary endpoint (current one)
		{
			URL:         "https://israelrail.azurefd.net/rjpa-prod/api/v1/timetable/searchTrainLuzForDateTime",
			Format:      "azure",
			RequiresKey: true,
		},
		// Alternative endpoint from research
		{
			URL:         "http://191.233.107.3/rail/v01/schedulev2/",
			Format:      "mobile",
			RequiresKey: false,
		},
		// Backup endpoint (mobile format)
		{
			URL:         "https://www.rail.co.il/rail/v01/schedulev2/",
			Format:      "mobile",
			RequiresKey: false,
		},
	}
	
	return &Service{
		config:    cfg,
		logger:    log,
		cache:     cache.NewCache(),
		client:    &http.Client{Timeout: 30 * time.Second},
		endpoints: endpoints,
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

	// Try each endpoint until one works
	for i, endpoint := range s.endpoints {
		s.logger.Info("Trying API endpoint",
			zap.Int("endpoint_index", i+1),
			zap.Int("total_endpoints", len(s.endpoints)),
			zap.String("url", endpoint.URL),
			zap.String("format", endpoint.Format))
		
		result, err := s.tryEndpoint(ctx, endpoint, from, to)
		if err == nil {
			s.logger.Info("Successfully got data from endpoint",
				zap.Int("endpoint_index", i+1),
				zap.String("url", endpoint.URL))
			// Cache result
			s.cache.Set("", from, to, result)
			return result, nil
		}
		
		s.logger.Warn("Endpoint failed, trying next",
			zap.Int("endpoint_index", i+1),
			zap.String("url", endpoint.URL),
			zap.String("error", err.Error()))
	}
	
	return nil, fmt.Errorf("all API endpoints failed")
}

func (s *Service) tryEndpoint(ctx context.Context, endpoint APIEndpoint, from, to string) ([]string, error) {
	var req *http.Request
	var err error
	
	if endpoint.Format == "azure" {
		req, err = s.buildAzureRequest(ctx, endpoint.URL, from, to)
	} else {
		req, err = s.buildMobileRequest(ctx, endpoint.URL, from, to)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	req.Header.Set("User-Agent", s.config.Train.UserAgent)
	if endpoint.RequiresKey {
		req.Header.Set("ocp-apim-subscription-key", s.config.Train.APIKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Try parsing as different response formats based on endpoint
	var result string
	if endpoint.Format == "azure" {
		var schedule ScheduleResponse
		if err := json.NewDecoder(resp.Body).Decode(&schedule); err != nil {
			return nil, fmt.Errorf("failed to decode azure response: %w", err)
		}
		result = s.formatSchedule(schedule)
	} else {
		// Try mobile format first
		body, _ := io.ReadAll(resp.Body)
		var mobileSchedule MobileScheduleResponse
		if err := json.Unmarshal(body, &mobileSchedule); err == nil {
			result = s.formatMobileSchedule(mobileSchedule)
		} else {
			// Fallback to azure format
			var schedule ScheduleResponse
			if err := json.Unmarshal(body, &schedule); err != nil {
				return nil, fmt.Errorf("failed to decode mobile response: %w", err)
			}
			result = s.formatSchedule(schedule)
		}
	}
	
	return splitMessage(result), nil
}

func (s *Service) buildAzureRequest(ctx context.Context, baseURL, from, to string) (*http.Request, error) {
	params := url.Values{}
	params.Add("fromStation", from)
	params.Add("toStation", to)
	params.Add("date", time.Now().Format("2006-01-02"))
	params.Add("hour", time.Now().Format("15:04:05"))
	params.Add("scheduleType", "2")
	params.Add("systemType", "1")
	params.Add("languageId", "Hebrew")

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = params.Encode()
	return req, nil
}

func (s *Service) buildMobileRequest(ctx context.Context, baseURL, from, to string) (*http.Request, error) {
	params := url.Values{}
	params.Add("origin", from)
	params.Add("destination", to)
	params.Add("date", time.Now().Format("02/01/2006 15:04"))
	params.Add("hours", "12")

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = params.Encode()
	return req, nil
}

func (s *Service) shouldRetry(statusCode int) bool {
	// Retry on server errors (5xx) and rate limiting (429)
	return statusCode >= 500 || statusCode == 429
}

func (s *Service) formatMobileSchedule(schedule MobileScheduleResponse) string {
	var result string
	count := 0
	for i, train := range schedule.Data {
		if count >= 5 {
			break
		}
		result += fmt.Sprintf("🚆 %d:\n", i+1)
		result += fmt.Sprintf("  🚂 1:\n")
		result += fmt.Sprintf("    עליה: %s (רציף %d)\n",
			s.GetStationName(fmt.Sprintf("%d", train.OriginStationId)), train.OriginPlatform)
		result += fmt.Sprintf("    זמן יציאת הרכבת: %s\n",
			s.formatTime(train.DepartureTime))
		result += fmt.Sprintf("    אל: %s (רציף %d)\n",
			s.GetStationName(fmt.Sprintf("%d", train.DestinationStationId)), train.DestinationPlatform)
		result += fmt.Sprintf("    זמן הגעה: %s\n",
			s.formatTime(train.ArrivalTime))
		result += "\n"
		count++
	}
	return result
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
