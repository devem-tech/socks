package resolver

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

const cacheCleanupInterval = 1 * time.Minute // Interval for cache cleanup checks

type cacheEntry struct {
	ip       net.IP
	expireAt time.Time
}

// Cache is a structure that represents a DNS cache with a time-to-live (TTL) for entries.
type Cache struct {
	m     metrics
	ttl   time.Duration
	cache sync.Map
}

// New initializes a new Cache instance with a given TTL for cached entries.
// It starts a background goroutine to periodically clean up expired entries.
func New(m metrics, ttl time.Duration) *Cache {
	cache := &Cache{
		m:   m,
		ttl: ttl,
	}

	// Run the background cleanup goroutine
	go cache.startCleaner()

	return cache
}

// Resolve performs DNS resolution for a given host and caches the result with TTL.
// If a valid cached entry exists, it returns the cached IP address without performing DNS lookup.
func (c *Cache) Resolve(host string) (net.IP, error) {
	log.Printf("dns: resolve: %s", host) // @fixme

	defer c.m.Timer(mResolve)()

	// Check the cache without locking
	if entry, found := c.cache.Load(host); found {
		cached := entry.(cacheEntry)

		// Return the cached IP if it's still valid
		if time.Now().Before(cached.expireAt) {
			c.m.Increment(mHits)
			return cached.ip, nil
		}
	}

	// Increment cache miss
	c.m.Increment(mMisses)

	// Perform DNS lookup
	ips, err := net.LookupIP(host)
	if err != nil || len(ips) == 0 {
		return nil, fmt.Errorf("resolve host %s: %w", host, err)
	}

	// Store the result in cache
	c.cache.Store(host, cacheEntry{
		ip:       ips[0],
		expireAt: time.Now().Add(c.ttl),
	})

	return ips[0], nil
}

// startCleaner is a background function that periodically removes expired entries from the cache.
// It runs every cacheCleanupInterval and checks each entry's expiration time, deleting any expired entries.
func (c *Cache) startCleaner() {
	ticker := time.NewTicker(cacheCleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		var entries, evictions int64

		c.cache.Range(func(key, value any) bool {
			entries++

			entry := value.(cacheEntry)

			// Delete the entry if it has expired
			if time.Now().After(entry.expireAt) {
				c.cache.Delete(key)

				evictions++ // Count evictions for expired entries
			}

			return true
		})

		c.m.Count(mEvictions, evictions)
		c.m.Gauge(mEntries, float64(entries))
	}
}
