package render

import (
	"container/list"
	"os"
	"sync"
)

const (
	// DefaultCacheEntries caps rendered messages.
	DefaultCacheEntries = 64
	// DefaultCacheBytes caps rendered markdown bytes (4 MiB).
	DefaultCacheBytes = 4 << 20
	// AsyncMarkdownBytes: larger assistant payloads render off the input loop.
	AsyncMarkdownBytes = 4096
	// TargetFPS is the spinner/header animation budget.
	TargetFPS = 30
	// InputLatencyBudgetMS is the p99 key→view budget (milliseconds).
	InputLatencyBudgetMS = 50
)

// AsyncOff disables background markdown (STELL_TUI_ASYNC=0).
func AsyncOff() bool {
	v := os.Getenv("STELL_TUI_ASYNC")
	return v == "0" || v == "false"
}

type cacheKey struct {
	id    int
	width int
}

type cacheEntry struct {
	key  cacheKey
	body string
}

// Cache is an LRU of rendered markdown keyed by message id + width.
type Cache struct {
	mu    sync.Mutex
	maxN  int
	maxB  int
	bytes int
	items map[cacheKey]*list.Element
	order *list.List
}

// NewCache bounds entries and total bytes. Zero uses defaults.
func NewCache(maxEntries, maxBytes int) *Cache {
	if maxEntries <= 0 {
		maxEntries = DefaultCacheEntries
	}
	if maxBytes <= 0 {
		maxBytes = DefaultCacheBytes
	}
	return &Cache{
		maxN:  maxEntries,
		maxB:  maxBytes,
		items: map[cacheKey]*list.Element{},
		order: list.New(),
	}
}

// Get returns a cached render.
func (c *Cache) Get(id, width int) (string, bool) {
	if c == nil {
		return "", false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	el, ok := c.items[cacheKey{id, width}]
	if !ok {
		return "", false
	}
	c.order.MoveToFront(el)
	return el.Value.(cacheEntry).body, true
}

// Put stores a render, evicting LRU entries to stay in budget.
func (c *Cache) Put(id, width int, body string) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	k := cacheKey{id, width}
	if el, ok := c.items[k]; ok {
		old := el.Value.(cacheEntry)
		c.bytes -= len(old.body)
		el.Value = cacheEntry{key: k, body: body}
		c.bytes += len(body)
		c.order.MoveToFront(el)
		c.evict()
		return
	}
	el := c.order.PushFront(cacheEntry{key: k, body: body})
	c.items[k] = el
	c.bytes += len(body)
	c.evict()
}

// Len is the number of cached frames (tests).
func (c *Cache) Len() int {
	if c == nil {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.order.Len()
}

// Bytes is the payload size (tests).
func (c *Cache) Bytes() int {
	if c == nil {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.bytes
}

func (c *Cache) evict() {
	for c.order.Len() > c.maxN || c.bytes > c.maxB {
		el := c.order.Back()
		if el == nil {
			return
		}
		ent := el.Value.(cacheEntry)
		c.order.Remove(el)
		delete(c.items, ent.key)
		c.bytes -= len(ent.body)
		if c.bytes < 0 {
			c.bytes = 0
		}
	}
}
