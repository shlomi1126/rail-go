package cache

import (
	"fmt"
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	cache1 := NewCache()
	cache2 := NewCache()
	
	if cache1 != cache2 {
		t.Error("NewCache() should return the same singleton instance")
	}
	
	if cache1.ttl != time.Duration(CACHE_TTL)*time.Second {
		t.Errorf("Cache TTL should be %d seconds, got %v", CACHE_TTL, cache1.ttl)
	}
}

func TestCacheSetAndGet(t *testing.T) {
	cache := NewCache()
	
	// Clear cache for test isolation
	cache.value = make(map[[20]byte]trainData)
	
	username := "testuser"
	from := "3700"
	to := "4600"
	testData := []string{"train1", "train2"}
	
	// Test Set
	cache.Set(username, from, to, testData)
	
	// Test Get
	result := cache.Get(username, from, to)
	if result == nil {
		t.Fatal("Expected cached data, got nil")
	}
	
	resultSlice, ok := result.([]string)
	if !ok {
		t.Fatal("Expected []string, got different type")
	}
	
	if len(resultSlice) != len(testData) {
		t.Errorf("Expected %d items, got %d", len(testData), len(resultSlice))
	}
	
	for i, item := range resultSlice {
		if item != testData[i] {
			t.Errorf("Expected %s at index %d, got %s", testData[i], i, item)
		}
	}
}

func TestCacheGetNonExistent(t *testing.T) {
	cache := NewCache()
	
	// Clear cache for test isolation
	cache.value = make(map[[20]byte]trainData)
	
	result := cache.Get("nonexistent", "from", "to")
	if result != nil {
		t.Error("Expected nil for non-existent cache key, got value")
	}
}

func TestCacheTTLExpiration(t *testing.T) {
	cache := NewCache()
	
	// Clear cache and set short TTL for testing
	cache.value = make(map[[20]byte]trainData)
	cache.ttl = 100 * time.Millisecond
	
	username := "testuser"
	from := "3700"
	to := "4600"
	testData := "test data"
	
	// Set data
	cache.Set(username, from, to, testData)
	
	// Should be available immediately
	result := cache.Get(username, from, to)
	if result == nil {
		t.Error("Expected cached data immediately after set")
	}
	
	// Wait for TTL to expire
	time.Sleep(150 * time.Millisecond)
	
	// Add new data to trigger cleanup
	cache.Set("cleanup", "trigger", "key", "value")
	
	// Should be nil after expiration and cleanup
	result = cache.Get(username, from, to)
	if result != nil {
		t.Error("Expected nil after TTL expiration, got value")
	}
}

func TestCacheKeyGeneration(t *testing.T) {
	cache := NewCache()
	
	// Clear cache for test isolation
	cache.value = make(map[[20]byte]trainData)
	
	// Test that different combinations create different cache entries
	cache.Set("user1", "from1", "to1", "data1")
	cache.Set("user1", "from1", "to2", "data2")
	cache.Set("user1", "from2", "to1", "data3")
	cache.Set("user2", "from1", "to1", "data4")
	
	// All should be different entries
	if len(cache.value) != 4 {
		t.Errorf("Expected 4 cache entries, got %d", len(cache.value))
	}
	
	// Verify each can be retrieved correctly
	tests := []struct {
		user, from, to, expected string
	}{
		{"user1", "from1", "to1", "data1"},
		{"user1", "from1", "to2", "data2"},
		{"user1", "from2", "to1", "data3"},
		{"user2", "from1", "to1", "data4"},
	}
	
	for _, test := range tests {
		result := cache.Get(test.user, test.from, test.to)
		if result != test.expected {
			t.Errorf("Expected %s for %s-%s-%s, got %v", test.expected, test.user, test.from, test.to, result)
		}
	}
}

func TestCacheCleanup(t *testing.T) {
	cache := NewCache()
	
	// Clear cache and set short TTL for testing
	cache.value = make(map[[20]byte]trainData)
	cache.ttl = 50 * time.Millisecond
	
	// Add multiple entries
	cache.Set("user1", "from1", "to1", "data1")
	cache.Set("user2", "from2", "to2", "data2")
	
	if len(cache.value) != 2 {
		t.Errorf("Expected 2 cache entries, got %d", len(cache.value))
	}
	
	// Wait for TTL to expire
	time.Sleep(60 * time.Millisecond)
	
	// Add new entry to trigger cleanup
	cache.Set("user3", "from3", "to3", "data3")
	
	// Only the new entry should remain
	if len(cache.value) != 1 {
		t.Errorf("Expected 1 cache entry after cleanup, got %d", len(cache.value))
	}
	
	// Verify the new entry is still accessible
	result := cache.Get("user3", "from3", "to3")
	if result != "data3" {
		t.Errorf("Expected 'data3', got %v", result)
	}
	
	// Verify old entries are gone
	result1 := cache.Get("user1", "from1", "to1")
	result2 := cache.Get("user2", "from2", "to2")
	if result1 != nil || result2 != nil {
		t.Error("Expected old entries to be cleaned up")
	}
}

func TestCacheConcurrentAccess(t *testing.T) {
	cache := NewCache()
	
	// Clear cache for test isolation
	cache.mutex.Lock()
	cache.value = make(map[[20]byte]trainData)
	cache.mutex.Unlock()
	
	const numGoroutines = 100
	const numOperations = 10
	
	// Test concurrent writes
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numOperations; j++ {
				cache.Set(fmt.Sprintf("user%d", id), fmt.Sprintf("from%d", j), fmt.Sprintf("to%d", j), fmt.Sprintf("data%d_%d", id, j))
			}
		}(i)
	}
	
	// Test concurrent reads
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numOperations; j++ {
				cache.Get(fmt.Sprintf("user%d", id), fmt.Sprintf("from%d", j), fmt.Sprintf("to%d", j))
			}
		}(i)
	}
	
	// Give goroutines time to complete
	time.Sleep(500 * time.Millisecond)
	
	// Test that cache is still functional after concurrent access
	cache.Set("test", "from", "to", "value")
	result := cache.Get("test", "from", "to")
	if result != "value" {
		t.Error("Cache not functional after concurrent access")
	}
}

func TestCacheRaceCondition(t *testing.T) {
	cache := NewCache()
	
	// Clear cache for test isolation
	cache.mutex.Lock()
	cache.value = make(map[[20]byte]trainData)
	cache.mutex.Unlock()
	
	const numGoroutines = 50
	done := make(chan bool, numGoroutines*2)
	
	// Start multiple goroutines that read and write simultaneously
	for i := 0; i < numGoroutines; i++ {
		// Writer goroutine
		go func(id int) {
			defer func() { done <- true }()
			for j := 0; j < 20; j++ {
				cache.Set("user", "from", "to", fmt.Sprintf("data%d_%d", id, j))
				time.Sleep(1 * time.Millisecond)
			}
		}(i)
		
		// Reader goroutine
		go func(id int) {
			defer func() { done <- true }()
			for j := 0; j < 20; j++ {
				cache.Get("user", "from", "to")
				time.Sleep(1 * time.Millisecond)
			}
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines*2; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out - possible deadlock")
		}
	}
	
	// Verify cache is still operational
	cache.Set("final", "test", "check", "success")
	result := cache.Get("final", "test", "check")
	if result != "success" {
		t.Error("Cache corrupted after race condition test")
	}
}