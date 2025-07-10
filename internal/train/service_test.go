package train

import (
	"testing"

	"rail-go/internal/logger"
)

func TestGetStationSuggestions(t *testing.T) {
	cfg := NewTestConfig()
	log := logger.New()
	service := NewService(cfg, log)

	tests := []struct {
		name     string
		query    string
		expected map[string]string
	}{
		{
			name:     "Hebrew station name",
			query:    "תל אביב",
			expected: map[string]string{"תל אביב - סבידור מרכז": "3700", "תל אביב - ההגנה": "4900", "תל אביב - אוניברסיטה": "3600", "תל אביב - השלום": "4600"},
		},
		{
			name:     "English station name",
			query:    "tel aviv",
			expected: map[string]string{"תל אביב - סבידור מרכז": "3700", "תל אביב - ההגנה": "4900", "תל אביב - אוניברסיטה": "3600", "תל אביב - השלום": "4600"},
		},
		{
			name:     "No matches",
			query:    "nonexistent",
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetStationSuggestions(tt.query)
			if len(result) != len(tt.expected) {
				t.Errorf("GetStationSuggestions() got %d results, want %d", len(result), len(tt.expected))
			}
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("GetStationSuggestions() for key %s got %s, want %s", k, result[k], v)
				}
			}
		})
	}
}

func TestGetStationName(t *testing.T) {
	cfg := NewTestConfig()
	log := logger.New()
	service := NewService(cfg, log)

	tests := []struct {
		name      string
		stationID string
		expected  string
	}{
		{
			name:      "Existing station",
			stationID: "3700",
			expected:  "תל אביב - סבידור מרכז",
		},
		{
			name:      "Non-existing station",
			stationID: "9999",
			expected:  "9999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetStationName(tt.stationID)
			if result != tt.expected {
				t.Errorf("GetStationName() got %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	cfg := NewTestConfig()
	log := logger.New()
	service := NewService(cfg, log)

	tests := []struct {
		name     string
		timeStr  string
		expected string
	}{
		{
			name:     "Valid time",
			timeStr:  "2024-04-21T15:30:00",
			expected: "15:30:00",
		},
		{
			name:     "Invalid time",
			timeStr:  "invalid",
			expected: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.formatTime(tt.timeStr)
			if result != tt.expected {
				t.Errorf("formatTime() got %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestSplitMessage(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected []string
	}{
		{
			name:    "Short message",
			message: "Hello\nWorld",
			expected: []string{
				"Hello\nWorld\n",
			},
		},
		{
			name:    "Empty message",
			message: "",
			expected: []string{
				"\n",
			},
		},
		{
			name:    "Single line",
			message: "Hello World",
			expected: []string{
				"Hello World\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitMessage(tt.message)
			if len(result) != len(tt.expected) {
				t.Errorf("splitMessage() got %d chunks, want %d", len(result), len(tt.expected))
			}
			for i, chunk := range result {
				if chunk != tt.expected[i] {
					t.Errorf("splitMessage() chunk %d got %q, want %q", i, chunk, tt.expected[i])
				}
			}
		})
	}
}

func TestCacheIntegration(t *testing.T) {
	cfg := NewTestConfig()
	log := logger.New()
	service := NewService(cfg, log)
	
	from := "3700"
	to := "4600"
	testData := []string{"cached train data"}
	
	// Set data directly in cache
	service.cache.Set("", from, to, testData)
	
	// Test that cached data is returned
	result := service.cache.Get("", from, to)
	if result == nil {
		t.Fatal("Expected cached data, got nil")
	}
	
	resultSlice, ok := result.([]string)
	if !ok {
		t.Fatal("Expected []string from cache, got different type")
	}
	
	if len(resultSlice) != len(testData) {
		t.Errorf("Expected %d cached items, got %d", len(testData), len(resultSlice))
	}
	
	for i, item := range resultSlice {
		if item != testData[i] {
			t.Errorf("Expected cached item %s at index %d, got %s", testData[i], i, item)
		}
	}
}

func TestCacheTTLIntegration(t *testing.T) {
	cfg := NewTestConfig()
	log := logger.New()
	service := NewService(cfg, log)
	
	from := "3700"
	to := "4600"
	testData := []string{"test data"}
	
	// Set data in cache
	service.cache.Set("", from, to, testData)
	
	// Should be available immediately
	result := service.cache.Get("", from, to)
	if result == nil {
		t.Error("Expected cached data immediately after set")
	}
	
	// Note: We can't easily test TTL expiration in integration test
	// since the cache TTL is set to 30 seconds by default.
	// The TTL functionality is thoroughly tested in cache_test.go
}
