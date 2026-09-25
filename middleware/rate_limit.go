package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	apperr "banking/pkg/errors"
	"banking/pkg/response"
)

type bucket struct {
	tokens float64
	last   time.Time
}

type limiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     float64 // token per detik
	capacity float64
}

// RateLimit membatasi jumlah request per IP menggunakan algoritma token bucket.
// Contoh: RateLimit(5, time.Minute) -> 5 request per menit per IP.
func RateLimit(requests int, per time.Duration) gin.HandlerFunc {
	l := &limiter{
		buckets:  make(map[string]*bucket),
		rate:     float64(requests) / per.Seconds(),
		capacity: float64(requests),
	}
	go l.cleanup()

	return func(c *gin.Context) {
		key := c.ClientIP() + "|" + c.FullPath()
		if !l.allow(key) {
			response.Error(c, apperr.ErrTooManyReq)
			return
		}
		c.Next()
	}
}

func (l *limiter) allow(key string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	b, ok := l.buckets[key]
	if !ok {
		l.buckets[key] = &bucket{tokens: l.capacity - 1, last: now}
		return true
	}

	b.tokens += now.Sub(b.last).Seconds() * l.rate
	if b.tokens > l.capacity {
		b.tokens = l.capacity
	}
	b.last = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func (l *limiter) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		threshold := time.Now().Add(-30 * time.Minute)
		l.mu.Lock()
		for key, b := range l.buckets {
			if b.last.Before(threshold) {
				delete(l.buckets, key)
			}
		}
		l.mu.Unlock()
	}
}
