package cache

import "sync"

type Cache struct {
	data map[string]cache_item
	mutex sync.RWMutex
}

const cache_size = 50

type cache_item struct {
	id int
	length int
	excerpt string
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[string]cache_item),
	}
}

func (c *Cache) Set(key string, value cache_item) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data[key] = value
}

func (c *Cache) Get(key string) (cache_item, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	value, exists := c.data[key]
	return value, exists
}

func (c *Cache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.data, key)
}
