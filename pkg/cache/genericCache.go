package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/DSSD-Madison/gmu/pkg/logger"
)

type GenericCache[T any] struct {
	items map[string]CacheItem[T]
	log   logger.Logger
	mu    sync.RWMutex
}

func NewGeneric[T any](log logger.Logger) *GenericCache[T] {
	cacheLogger := log.With("cache", fmt.Sprintf("%T", *new(T)))
	return &GenericCache[T]{
		items: make(map[string]CacheItem[T]),
		log:   cacheLogger,
	}
}

func (c *GenericCache[T]) Set(key string, value T, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiration := time.Now().Add(duration).Unix()
	c.log.Info("adding value to cache", "key", key, "expires", expiration)
	c.items[key] = CacheItem[T]{
		Value:      value,
		Expiration: expiration,
	}
}

func (c *GenericCache[T]) Get(key string) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		var zero T
		return zero, false
	}

	if time.Now().Unix() > item.Expiration && item.Expiration != 0 {
		c.delete(key)
		var zero T
		return zero, false
	}

	return item.Value, true
}

func (c *GenericCache[T]) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.delete(key)
}

func (c *GenericCache[T]) delete(key string) {
	delete(c.items, key)
}

func (c *GenericCache[T]) CleanExpiredItems(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			for key, item := range c.items {
				if time.Now().Unix() > item.Expiration && item.Expiration != 0 {
					delete(c.items, key)
				}
			}
			c.mu.Unlock()
		case <-ctx.Done():
			c.log.DebugContext(ctx, "Cache expiration cleaner shutting down")
			return
		}
	}
}
