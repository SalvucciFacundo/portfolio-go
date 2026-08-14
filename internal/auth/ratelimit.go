package auth

import (
	"sync"
	"time"
)

const (
	rateLimitWindow  = time.Minute
	rateLimitMaxHits = 5
)

// Limiter is an in-memory per-IP rate limiter with a sliding window. It is
// safe for concurrent use.
type Limiter struct {
	mu   sync.Mutex
	hits map[string][]time.Time
}

// NewLimiter creates an empty Limiter.
func NewLimiter() *Limiter {
	return &Limiter{hits: make(map[string][]time.Time)}
}

// Allow reports whether ip may perform one more action within the current
// window. Stale hits older than the window are pruned lazily.
func (l *Limiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := time.Now().Add(-rateLimitWindow)
	recent := l.hits[ip][:0]
	for _, t := range l.hits[ip] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}

	if len(recent) >= rateLimitMaxHits {
		l.hits[ip] = recent
		return false
	}

	l.hits[ip] = append(recent, time.Now())
	return true
}
