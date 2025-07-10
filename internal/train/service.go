package train

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"rail-go/internal/cache"
	"rail-go/internal/config"
	"rail-go/internal/logger"

	"go.uber.org/zap"
)

type Service struct {
	config    *config.Config
	logger    *logger.Logger
	cache     cache.CacheInterface
	client    HTTPClient
	endpoints []Endpoint
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
	endpoints := []Endpoint{
		// Primary endpoint (current one)
		NewAzureEndpoint("https://israelrail.azurefd.net/rjpa-prod/api/v1/timetable/searchTrainLuzForDateTime", cfg),
		// Alternative endpoint from research
		NewMobileEndpoint("http://191.233.107.3/rail/v01/schedulev2/", cfg),
		// Backup endpoint (mobile format)
		NewMobileEndpoint("https://www.rail.co.il/rail/v01/schedulev2/", cfg),
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

func (s *Service) getScheduleWithRetry(ctx context.Context, from, to string, _ int) ([]string, error) {
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
			zap.String("url", endpoint.GetURL()),
			zap.String("format", endpoint.GetFormat()))

		result, err := s.tryEndpoint(ctx, endpoint, from, to)
		if err == nil {
			s.logger.Info("Successfully got data from endpoint",
				zap.Int("endpoint_index", i+1),
				zap.String("url", endpoint.GetURL()))
			// Cache result
			s.cache.Set("", from, to, result)
			return result, nil
		}

		s.logger.Warn("Endpoint failed, trying next",
			zap.Int("endpoint_index", i+1),
			zap.String("url", endpoint.GetURL()),
			zap.String("error", err.Error()))
	}

	return nil, fmt.Errorf("all API endpoints failed")
}

func (s *Service) tryEndpoint(ctx context.Context, endpoint Endpoint, from, to string) ([]string, error) {
	req, err := endpoint.BuildRequest(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", s.config.Train.UserAgent)
	if endpoint.RequiresAPIKey() {
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

	// Parse response using endpoint-specific parser
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	result, err := endpoint.ParseResponse(body)
	if err != nil {
		return nil, err
	}

	return splitMessage(result), nil
}

func (s *Service) GetStationName(stationID string) string {
	return getStationName(stationID)
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
