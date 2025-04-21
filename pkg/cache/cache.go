package cache

import (
	"context"
	"time"
)

type CacheItem[T any] struct {
	Value      T
	Expiration int64
}

type Cache[T any] interface {
	Set(key string, value T, duration time.Duration)
	Get(key string) (T, bool)
	Delete(key string)
	CleanExpiredItems(ctx context.Context, interval time.Duration)
}
