package user

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type cacheItem struct {
	data     any
	expireAt time.Time
}

type cache struct {
	mu    sync.Mutex
	items map[uuid.UUID]cacheItem
}

func newCache() *cache {
	return &cache{
		items: make(map[uuid.UUID]cacheItem),
	}
}

func (c *cache) set(key uuid.UUID, data any, expireAt time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = cacheItem{
		data:     data,
		expireAt: time.Now().Add(expireAt),
	}
}

func (c *cache) get(key uuid.UUID) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.items[key]
	if !ok {
		return nil, false
	}

	if item.expireAt.Before(time.Now()) {
		return nil, false
	}

	return item.data, true
}
