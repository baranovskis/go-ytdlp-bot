package cache

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// Result holds the cached download result.
type Result struct {
	FilePath string
	Filename string
	Title    string
}

// DownloadFunc performs the actual download and returns the result.
type DownloadFunc func(ctx context.Context) (*Result, error)

type entry struct {
	result *Result
	ready  chan struct{}
	err    error
	expAt  time.Time
}

// Cache coordinates concurrent downloads of the same URL and provides
// TTL-based file cleanup.
type Cache struct {
	mu          sync.Mutex
	entries     map[string]*entry
	ttl         time.Duration
	removeFiles bool
	logger      zerolog.Logger
	stopCleanup chan struct{}
}

// New creates a new download cache.
func New(ttl time.Duration, removeFiles bool, logger zerolog.Logger) *Cache {
	c := &Cache{
		entries:     make(map[string]*entry),
		ttl:         ttl,
		removeFiles: removeFiles,
		logger:      logger,
		stopCleanup: make(chan struct{}),
	}
	go c.cleanupLoop()
	return c
}

// GetOrDownload returns a cached result or runs downloadFn exactly once per URL.
// Concurrent callers for the same URL will wait for the single download to complete.
func (c *Cache) GetOrDownload(ctx context.Context, url string, downloadFn DownloadFunc) (*Result, error) {
	c.mu.Lock()

	if e, ok := c.entries[url]; ok {
		select {
		case <-e.ready:
			// Download complete — check if still valid
			if e.err == nil && time.Now().Before(e.expAt) {
				c.mu.Unlock()
				return e.result, nil
			}
			// Expired or errored — remove and re-download
			delete(c.entries, url)
		default:
			// Download in progress — wait for it
			c.mu.Unlock()
			select {
			case <-e.ready:
				return e.result, e.err
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	// Create new entry
	e := &entry{
		ready: make(chan struct{}),
	}
	c.entries[url] = e
	c.mu.Unlock()

	// Perform download
	result, err := downloadFn(ctx)
	e.result = result
	e.err = err
	if err == nil {
		e.expAt = time.Now().Add(c.ttl)
	}
	close(e.ready)

	if err != nil {
		c.mu.Lock()
		delete(c.entries, url)
		c.mu.Unlock()
	}

	return result, err
}

func (c *Cache) cleanupLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopCleanup:
			return
		}
	}
}

func (c *Cache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for url, e := range c.entries {
		select {
		case <-e.ready:
			if now.After(e.expAt) {
				if c.removeFiles && e.result != nil {
					if err := os.Remove(e.result.FilePath); err != nil && !os.IsNotExist(err) {
						c.logger.Error().
							Str("path", e.result.FilePath).
							Str("error", err.Error()).
							Msg("failed to remove cached file")
					} else if err == nil {
						c.logger.Debug().
							Str("path", e.result.FilePath).
							Msg("removed expired cached file")
					}
				}
				delete(c.entries, url)
			}
		default:
			// Download in progress, skip
		}
	}
}

// Stop stops the cleanup goroutine.
func (c *Cache) Stop() {
	close(c.stopCleanup)
}
