package main

import (
	"sync"
	"time"
)

// entry is one element in the cache
type entry struct {
	b  []byte
	ts time.Time
}

// cache is a max sized cache with expiry.
type cache struct {
	size   int
	expiry time.Duration
	mu     sync.Mutex
	m      map[string]entry
}

func newCache(size int, expiry time.Duration) *cache {
	if size < 1 {
		panic("zero")
	}
	return &cache{
		size:   size,
		expiry: expiry,
		m:      make(map[string]entry),
	}
}

func (c *cache) addAt(k string, b []byte, ts time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// prune if we're oversize (right now random, but should be oldest)
	for len(c.m) >= c.size {
		for kd := range c.m {
			delete(c.m, kd)
			break
		}
	}

	c.m[k] = entry{b: b, ts: ts}
}

func (c *cache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.m)
}

func (c *cache) Add(k string, b []byte) {
	c.addAt(k, b, time.Now().UTC())
}

func (c *cache) Get(k string) []byte {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.m[k]
	if !ok {
		return nil
	}

	now := time.Now().UTC()
	cutoff := now.Add(-c.expiry)
	if e.ts.Before(cutoff) {
		delete(c.m, k)
		return nil
	}
	return e.b
}

func (c *cache) Keys() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	keys := make([]string, 0, len(c.m))
	for k := range c.m {
		keys = append(keys, k)
	}
	return keys
}
