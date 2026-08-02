package pokecache

import (
	"sync"
	"time"
)

type cacheItem struct {
	value     []byte
	createdAt time.Time
}
type Cache struct {
	items        map[string]cacheItem
	mu           sync.RWMutex
	stopChan     chan struct{}
	cleanUpInterval time.Duration
}

func NewCache(cleanUpInterval time.Duration) *Cache {
	c := &Cache{
		items:        make(map[string]cacheItem),
		stopChan:     make(chan struct{}),
		cleanUpInterval : cleanUpInterval,
	}
	go c.reapLoop()
	return c
}
func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = cacheItem{
		value:     val,
		createdAt: time.Now(),
	}
}
func (c *Cache) Get(key string) ([]byte, bool) {
	interval := c.cleanUpInterval
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, found := c.items[key]
	if !found || time.Now().After(item.createdAt.Add(interval)) {
		var zero []byte
		return zero, false
	}
	return item.value, true
}
func (c *Cache) reapLoop() {
	interval := c.cleanUpInterval
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			for key, _ := range c.items {
				item, found := c.items[key]
				if !found || time.Now().After(item.createdAt.Add(interval)) {
					delete(c.items, key)
				}
			}
			c.mu.Unlock()
		case <-c.stopChan:
			return
		}
	}
}

func (c *Cache) Close() {
	close(c.stopChan)
}
