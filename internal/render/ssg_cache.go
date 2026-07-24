package render

import (
	"sync"
	"time"
)

// ssgEntry is one cached, pre-rendered "ssg" page.
type ssgEntry struct {
	html       string
	renderedAt time.Time
	revalidate time.Duration // <= 0 means "never automatically revalidate"

	mu           sync.Mutex
	revalidating bool
}

// stale reports whether entry is old enough that it should be
// regenerated in the background before being served again.
func (e *ssgEntry) stale() bool {
	if e.revalidate <= 0 {
		return false
	}
	return time.Since(e.renderedAt) >= e.revalidate
}

// beginRevalidate atomically claims the right to run a background
// revalidation for this entry, returning false if one is already running.
func (e *ssgEntry) beginRevalidate() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.revalidating {
		return false
	}
	e.revalidating = true
	return true
}

func (e *ssgEntry) endRevalidate() {
	e.mu.Lock()
	e.revalidating = false
	e.mu.Unlock()
}

// ssgCache holds pre-rendered "ssg" page output in memory, keyed by route.
//
// It implements the "revalidate: N" ISR-style pattern from
// 03-RENDERING-ENGINE.md: a request arriving after an entry has gone
// stale is still served the (stale) cached HTML immediately for speed,
// while a fresh render is kicked off in the background; only the *next*
// request after that background render completes sees the updated
// content.
type ssgCache struct {
	mu      sync.RWMutex
	entries map[string]*ssgEntry
}

func newSSGCache() *ssgCache {
	return &ssgCache{entries: make(map[string]*ssgEntry)}
}

func (c *ssgCache) set(route, html string, revalidate time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[route] = &ssgEntry{html: html, renderedAt: time.Now(), revalidate: revalidate}
}

func (c *ssgCache) get(route string) (*ssgEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[route]
	return e, ok
}
