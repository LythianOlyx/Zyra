package dx

import (
	"container/list"
	"context"
	"sync"
	"time"
)

type cacheItem struct {
	key       string
	value     any
	expiresAt time.Time
}

// MemoryLRUCache is a high-performance in-memory LRU cache with item TTL expiration.
type MemoryLRUCache struct {
	mu       sync.Mutex
	capacity int
	items    map[string]*list.Element
	evictList *list.List
}

func NewMemoryLRUCache(capacity int) *MemoryLRUCache {
	if capacity <= 0 {
		capacity = 10000
	}
	return &MemoryLRUCache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		evictList: list.New(),
	}
}

func (c *MemoryLRUCache) Set(key string, value any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}

	if elem, exists := c.items[key]; exists {
		c.evictList.MoveToFront(elem)
		item := elem.Value.(*cacheItem)
		item.value = value
		item.expiresAt = exp
		return
	}

	item := &cacheItem{
		key:       key,
		value:     value,
		expiresAt: exp,
	}
	elem := c.evictList.PushFront(item)
	c.items[key] = elem

	if c.evictList.Len() > c.capacity {
		c.evictOldest()
	}
}

func (c *MemoryLRUCache) Get(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, exists := c.items[key]
	if !exists {
		return nil, false
	}

	item := elem.Value.(*cacheItem)
	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		c.removeElement(elem)
		return nil, false
	}

	c.evictList.MoveToFront(elem)
	return item.value, true
}

func (c *MemoryLRUCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.items[key]; exists {
		c.removeElement(elem)
	}
}

func (c *MemoryLRUCache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*list.Element)
	c.evictList.Init()
}

func (c *MemoryLRUCache) evictOldest() {
	elem := c.evictList.Back()
	if elem != nil {
		c.removeElement(elem)
	}
}

func (c *MemoryLRUCache) removeElement(elem *list.Element) {
	c.evictList.Remove(elem)
	item := elem.Value.(*cacheItem)
	delete(c.items, item.key)
}

// CacheManager handles cache operations with generic support for Remember.
type CacheManager struct {
	store *MemoryLRUCache
}

func NewCacheManager() *CacheManager {
	return &CacheManager{
		store: NewMemoryLRUCache(10000),
	}
}

func (m *CacheManager) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
	m.store.Set(key, val, ttl)
	return nil
}

func (m *CacheManager) Get(ctx context.Context, key string) (any, bool, error) {
	val, ok := m.store.Get(key)
	return val, ok, nil
}

func (m *CacheManager) Delete(ctx context.Context, key string) error {
	m.store.Delete(key)
	return nil
}

func (m *CacheManager) Flush(ctx context.Context) error {
	m.store.Flush()
	return nil
}

// RememberGeneric returns cached value or executes fn and caches result.
func RememberGeneric[T any](ctx context.Context, mgr *CacheManager, key string, ttl time.Duration, fn func() (T, error)) (T, error) {
	if val, ok, _ := mgr.Get(ctx, key); ok {
		if typed, valid := val.(T); valid {
			return typed, nil
		}
	}

	res, err := fn()
	if err != nil {
		var zero T
		return zero, err
	}

	_ = mgr.Set(ctx, key, res, ttl)
	return res, nil
}
