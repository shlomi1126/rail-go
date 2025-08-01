package cache

// CacheInterface defines the contract for cache implementations
type CacheInterface interface {
	Get(username, from, to string) any
	Set(username, from, to string, value any)
}
