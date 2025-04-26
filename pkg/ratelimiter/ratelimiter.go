package ratelimiter

// RateLimiter defines the interface for rate limiting a service
type RateLimiter interface {
	// IsLimited checks if the given key is currently rate-limited.
	// It returns true if limited, false otherwise.
	IsLimited(key string) bool

	//RecordAttempt records an access attempt for the given key.
	// 'success' should be true for a successful access, false for a failed access.
	// This method should update the internal state of the rate limiter,
	// potentially triggering a limit if the attempt fails and thresholds are met.
	RecordAttempt(key string, success bool)
}
