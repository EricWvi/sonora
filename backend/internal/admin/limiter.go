package admin

import (
	"sync"
	"time"
)

// limiter refuses a key after max failures within a fixed window. Keys
// are not per client: behind a reverse proxy every client shares an
// address, and the secrets it protects are long enough that slowing every
// guesser down together is the point.
type limiter struct {
	max    int
	window time.Duration

	mu      sync.Mutex
	windows map[string]*failures
}

type failures struct {
	start time.Time
	count int
}

func newLimiter(max int, window time.Duration) *limiter {
	return &limiter{max: max, window: window, windows: map[string]*failures{}}
}

// allow reports whether key may try again, and if not, how long to wait.
func (l *limiter) allow(key string, now time.Time) (time.Duration, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	f := l.windows[key]
	if f == nil || now.Sub(f.start) >= l.window {
		return 0, true
	}
	if f.count >= l.max {
		return f.start.Add(l.window).Sub(now), false
	}
	return 0, true
}

// fail records a failed attempt.
func (l *limiter) fail(key string, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	f := l.windows[key]
	if f == nil || now.Sub(f.start) >= l.window {
		f = &failures{start: now}
		l.windows[key] = f
	}
	f.count++
}
