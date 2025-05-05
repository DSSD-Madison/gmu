package ratelimiter

import (
	"sync"
	"time"

	"github.com/DSSD-Madison/gmu/pkg/logger"
)

// InMemoryRateLimiter is a simple rate limiter that stores state in memory.
// It limits based on failed attempts within a sliding window.
type InMemoryRateLimiter struct {
	maxFailedAttempts int           // Max failed attempts before limiting
	blockDuration     time.Duration // How long to block after exceeding the limit
	attemptWindow     time.Duration // Time window for counting failed attempts

	failedAttempts map[string]int
	firstAttempt   map[string]time.Time // Timestamp of the first failed attempt in the current window
	blockUntil     map[string]time.Time // Timestamp until the key is blocked

	mu       sync.Mutex
	stopChan chan struct{}  // Channel to signal cleaner to stop
	wg       sync.WaitGroup // Waitgroup to wati for cleaner goroutine
	log      logger.Logger
}

// NewInMemoryRateLimiter creates a new instance of InMemoryRateLimiter.
func NewInMemoryRateLimiter(log logger.Logger, maxAttempts int, blockDur, windowDur time.Duration) *InMemoryRateLimiter {
	limiter := &InMemoryRateLimiter{
		maxFailedAttempts: maxAttempts,
		blockDuration:     blockDur,
		attemptWindow:     windowDur,
		failedAttempts:    make(map[string]int),
		firstAttempt:      make(map[string]time.Time),
		blockUntil:        make(map[string]time.Time),
		stopChan:          make(chan struct{}),
		log:               log.With("component", "InMemoryRateLimiter"),
	}

	// Start a background cleaner to remove expired entries
	limiter.wg.Add(1)
	go limiter.cleaner()

	return limiter
}

// IsLimited checks if the key is currently blocked.
func (l *InMemoryRateLimiter) IsLimited(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	// If successful, clear any existing state for this key
	blockTime, ok := l.blockUntil[key]
	return ok && time.Now().Before(blockTime)
}

// RecordAttempt records an access attempt and updates the rate limit state.
func (l *InMemoryRateLimiter) RecordAttempt(key string, success bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// If successful, clear any existing state for this key
	if success {
		delete(l.failedAttempts, key)
		delete(l.firstAttempt, key)
		delete(l.blockUntil, key)
	}

	// --- Failed attempt logic ---

	now := time.Now()

	// If the window for this key has expired since the first attempt, reset the count
	if firstAttemptTime, ok := l.firstAttempt[key]; ok && now.After(firstAttemptTime.Add(l.attemptWindow)) {
		l.failedAttempts[key] = 0
		delete(l.firstAttempt, key)
	}

	// If this is the first failed attempt in a new window for this key, record the time
	if _, ok := l.firstAttempt[key]; !ok {
		l.firstAttempt[key] = now
	}

	l.failedAttempts[key]++

	// If the count exceeds the limit, block the key
	if l.failedAttempts[key] >= l.maxFailedAttempts {
		l.blockUntil[key] = now.Add(l.blockDuration)
		l.log.Warn("Rate limit threshold exceeded, blocking key", "key", key, "attempts", l.failedAttempts[key], "block_until", now.Add(l.blockDuration))
	}
}

func (l *InMemoryRateLimiter) cleaner() {
	defer l.wg.Done()

	tickerBlocked := time.NewTicker(l.blockDuration / 2)
	defer tickerBlocked.Stop()

	tickerWindow := time.NewTicker(l.attemptWindow / 2)
	defer tickerWindow.Stop()

	for {
		select {
		case <-tickerBlocked.C:
			l.mu.Lock()
			now := time.Now()
			for key, blockTime := range l.blockUntil {
				if now.After(blockTime) {
					delete(l.blockUntil, key)
					delete(l.failedAttempts, key)
					delete(l.firstAttempt, key)
				}
			}
			l.mu.Unlock()
		case <-tickerWindow.C:
			l.mu.Lock()
			now := time.Now()
			for key, firstAttemptTime := range l.firstAttempt {
				// If the window expired and the key is not blocked (or block also expired)
				if now.After(firstAttemptTime.Add(l.attemptWindow)) {
					// Only clear if not currently blocked (a blocked key might still have failed attempts recorded)
					if _, isBlocked := l.blockUntil[key]; !isBlocked {
						delete(l.failedAttempts, key)
						delete(l.firstAttempt, key)
					}
				}
			}
			l.mu.Unlock()
		case <-l.stopChan:
			l.log.Info("Rate limiter cleaner received stop signal, exiting")
			return
		}
	}
}

// Shutdown signals the cleaner goroutine to stop and waits for it.
func (l *InMemoryRateLimiter) Shutdown() {
	close(l.stopChan)
	l.wg.Wait()
}
