package subscription

import "sync"

type Cache struct {
	mu sync.RWMutex

	data map[string]map[string]struct{}
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[string]map[string]struct{}),
	}
}

func (c *Cache) AddItems(key string, values ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, ok := c.data[key]
	if !ok {
		data = make(map[string]struct{})
	}

	for _, val := range values {
		data[val] = struct{}{}
	}

	c.data[key] = data
}

func (c *Cache) UpdateItems(key string, values ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	data := make(map[string]struct{})
	for _, val := range values {
		data[val] = struct{}{}
	}

	c.data[key] = data
}

func (c *Cache) GetItems(key string) ([]string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, ok := c.data[key]
	if !ok {
		return []string{}, false
	}

	res := make([]string, len(data))
	idx := 0
	for val := range data {
		res[idx] = val
		idx++
	}

	return res, true
}

func (c *Cache) RemoveItem(key, value string) {
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
