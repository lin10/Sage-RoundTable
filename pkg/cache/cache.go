package cache

import (
	"sync"
	"time"
)

// Item 缓存项
type Item struct {
	Value     interface{}
	ExpiresAt time.Time
}

// IsExpired 检查缓存项是否过期
func (item *Item) IsExpired() bool {
	return time.Now().After(item.ExpiresAt)
}

// Cache 简单的内存缓存实现
type Cache struct {
	mu    sync.RWMutex
	items map[string]*Item
	ttl   time.Duration
}

// NewCache 创建新的缓存实例
func NewCache(ttl time.Duration) *Cache {
	c := &Cache{
		items: make(map[string]*Item),
		ttl:   ttl,
	}
	
	// 启动定期清理任务
	go c.cleanupLoop()
	
	return c
}

// Set 设置缓存项
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.items[key] = &Item{
		Value:     value,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// SetWithTTL 设置带自定义 TTL 的缓存项
func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.items[key] = &Item{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// Get 获取缓存项
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	item, exists := c.items[key]
	if !exists {
		return nil, false
	}
	
	if item.IsExpired() {
		// 延迟删除，避免在读锁中修改
		go c.Delete(key)
		return nil, false
	}
	
	return item.Value, true
}

// Delete 删除缓存项
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	delete(c.items, key)
}

// Clear 清空所有缓存
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.items = make(map[string]*Item)
}

// Count 返回缓存项数量
func (c *Cache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return len(c.items)
}

// cleanupLoop 定期清理过期缓存项
func (c *Cache) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		c.cleanup()
	}
}

// cleanup 清理所有过期的缓存项
func (c *Cache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	now := time.Now()
	for key, item := range c.items {
		if now.After(item.ExpiresAt) {
			delete(c.items, key)
		}
	}
}
