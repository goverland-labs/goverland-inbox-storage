package subscription

import (
	"sync"

	"github.com/google/uuid"
)

type Cache struct {
	mu sync.RWMutex

	data map[string]map[uuid.UUID]struct{}
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[string]map[uuid.UUID]struct{}),
	}
}

func (c *Cache) AddItems(key string, values ...uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, ok := c.data[key]
	if !ok {
		data = make(map[uuid.UUID]struct{})
	}

	for _, val := range values {
		data[val] = struct{}{}
	}

	c.data[key] = data
}

func (c *Cache) UpdateItems(key string, values ...uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	data := make(map[uuid.UUID]struct{})
	for _, val := range values {
		data[val] = struct{}{}
	}

	c.data[key] = data
}

func (c *Cache) GetItems(key string) ([]uuid.UUID, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, ok := c.data[key]
	if !ok {
		return []uuid.UUID{}, false
	}

	res := make([]uuid.UUID, len(data))
	idx := 0
	for val := range data {
		res[idx] = val
		idx++
	}

	return res, true
}

func (c *Cache) RemoveItem(key string, value uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, ok := c.data[key]
	if !ok {
		return
	}

	delete(data, value)
}

func (c *Cache) RemoveKey(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
}
