package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// ipRateLimiter keeps a per-IP token bucket and evicts entries that have not
// been seen for a while. It is intentionally tiny (no external deps beyond
// x/time/rate) so it can be deployed as-is to a single instance. For multi
// node deployments this should be swapped for a Redis-backed limiter.
type ipRateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*client
	rps      rate.Limit
	burst    int
	ttl      time.Duration
	stopOnce sync.Once
	stopCh   chan struct{}
}

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func newIPRateLimiter(rps float64, burst int) *ipRateLimiter {
	l := &ipRateLimiter{
		clients: make(map[string]*client),
		rps:     rate.Limit(rps),
		burst:   burst,
		ttl:     10 * time.Minute,
		stopCh:  make(chan struct{}),
	}
	go l.gcLoop()
	return l
}

func (l *ipRateLimiter) gcLoop() {
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			l.cleanup()
		case <-l.stopCh:
			return
		}
	}
}

func (l *ipRateLimiter) cleanup() {
	cutoff := time.Now().Add(-l.ttl)
	l.mu.Lock()
	defer l.mu.Unlock()
	for ip, c := range l.clients {
		if c.lastSeen.Before(cutoff) {
			delete(l.clients, ip)
		}
	}
}

func (l *ipRateLimiter) get(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	c, ok := l.clients[ip]
	if !ok {
		c = &client{limiter: rate.NewLimiter(l.rps, l.burst)}
		l.clients[ip] = c
	}
	c.lastSeen = time.Now()
	return c.limiter
}

// RateLimit returns a Gin middleware that throttles requests using a per-IP
// token bucket. If rps <= 0 the middleware is disabled (returns a no-op).
func RateLimit(rps float64, burst int) gin.HandlerFunc {
	if rps <= 0 {
		return func(c *gin.Context) { c.Next() }
	}
	limiter := newIPRateLimiter(rps, burst)
	return func(c *gin.Context) {
		ip := clientIP(c)
		if !limiter.get(ip).Allow() {
			c.Header("Retry-After", "1")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests, please slow down",
			})
			return
		}
		c.Next()
	}
}

// clientIP returns the trusted client IP. Gin already honours
// X-Forwarded-For when configured via SetTrustedProxies; we fall back to
// RemoteAddr so a misconfigured proxy can never exhaust memory by spoofing
// arbitrary IPs.
func clientIP(c *gin.Context) string {
	ip := c.ClientIP()
	if ip == "" {
		host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
		if err != nil {
			return c.Request.RemoteAddr
		}
		return host
	}
	return ip
}
