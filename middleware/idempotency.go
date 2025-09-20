package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/EricWvi/sonora/log"
	"github.com/gin-gonic/gin"
)

// Simple in-memory cache (like Redis)
type cacheEntry struct {
	data      []byte
	status    int
	createdAt time.Time
}

var (
	idemCache = make(map[string]*cacheEntry)
	mu        sync.RWMutex
)

func init() {
	// start cleanup worker
	startCacheCleaner(1 * time.Minute)
}

func readCache(key string) (*cacheEntry, bool) {
	mu.RLock()
	defer mu.RUnlock()
	entry, ok := idemCache[key]
	return entry, ok
}

// firstWriteCache returns isFirstWrite
func firstWriteCache(key string, value *cacheEntry) bool {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := idemCache[key]; ok {
		return false
	} else {
		idemCache[key] = value
		return true
	}
}

func checkCache(c *gin.Context, key string) {
	start := time.Now()
	if entry, ok := readCache(key); ok {
		for time.Since(start) < 5*time.Second {
			if entry.status != -1 {
				log.Info(c, "idem cache hit")
				c.Data(entry.status, "application/json", entry.data)
				c.Abort()
				return
			}
			time.Sleep(500 * time.Millisecond)
		}
		log.Error(c, "read idem cache timeout")
		c.Data(http.StatusInternalServerError, "text/plain", []byte("read idem cache timeout"))
		c.Abort()
		return
	}
}

// Background job to evict expired keys
func startCacheCleaner(interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			count := 0
			mu.Lock()
			for k, v := range idemCache {
				if time.Since(v.createdAt) > 5*time.Minute {
					delete(idemCache, k)
					count++
				}
			}
			mu.Unlock()
			if count > 0 {
				log.Debugf(log.IdemCtx, "idemCache clean %d keys", count)
			}
		}
	}()
}

func Idempotency() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("Idempotency-Key")
		if key == "" {
			c.Next()
			return
		}

		c.Set(log.RequestIDCtxKey, key)

		// Check cache
		checkCache(c, key)
		if c.IsAborted() {
			return
		}

		isFirstWrite := firstWriteCache(key, &cacheEntry{
			status:    -1,
			createdAt: time.Now(),
		}) // placeholder
		if !isFirstWrite {
			checkCache(c, key)
			return
		}

		c.Next()

		mu.Lock()
		defer mu.Unlock()
		if entry, ok := idemCache[key]; ok {
			if c.IsAborted() {
				entry.data = []byte(`"previous request failed"`)
				entry.status = http.StatusInternalServerError
			} else {
				writer, _ := c.Get("bodyWriter")
				rspBody := writer.(*bodyWriter).body.Bytes()
				entry.data = rspBody
				entry.status = c.Writer.Status()
				entry.createdAt = time.Now()
			}
		}
	}
}
