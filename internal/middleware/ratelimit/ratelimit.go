package ratelimit

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Algorithm defines the rate limiting algorithm
type Algorithm string

const (
	TokenBucket   Algorithm = "token_bucket"
	SlidingWindow Algorithm = "sliding_window"
	FixedWindow   Algorithm = "fixed_window"
)

// KeyExtractor defines how to extract the rate limiting key from a request
type KeyExtractor string

const (
	ExtractorIP     KeyExtractor = "ip"
	ExtractorHeader KeyExtractor = "header"
	ExtractorCustom KeyExtractor = "custom"
)

// Config holds rate limiting configuration
type Config struct {
	Enabled       bool
	Requests      int
	Window        time.Duration
	Burst         int
	Algorithm     Algorithm
	KeyExtractor  KeyExtractor
	HeaderName    string
	CustomKeyFunc func(*http.Request) string
}

// Limiter represents a rate limiter instance
type Limiter interface {
	Allow(key string) (bool, time.Duration)
	Reset(key string)
}

// TokenBucketLimiter implements token bucket algorithm
type TokenBucketLimiter struct {
	mu       sync.RWMutex
	buckets  map[string]*tokenBucket
	rate     float64
	capacity int
	window   time.Duration
}

type tokenBucket struct {
	tokens   float64
	lastSeen time.Time
}

// NewTokenBucketLimiter creates a new token bucket limiter
func NewTokenBucketLimiter(requests int, window time.Duration, burst int) *TokenBucketLimiter {
	rate := float64(requests) / window.Seconds()
	capacity := burst
	if capacity <= 0 {
		capacity = requests
	}
	
	return &TokenBucketLimiter{
		buckets:  make(map[string]*tokenBucket),
		rate:     rate,
		capacity: capacity,
		window:   window,
	}
}

func (tbl *TokenBucketLimiter) Allow(key string) (bool, time.Duration) {
	tbl.mu.Lock()
	defer tbl.mu.Unlock()
	
	now := time.Now()
	bucket, exists := tbl.buckets[key]
	
	if !exists {
		bucket = &tokenBucket{
			tokens:   float64(tbl.capacity),
			lastSeen: now,
		}
		tbl.buckets[key] = bucket
	}
	
	// Add tokens based on elapsed time
	elapsed := now.Sub(bucket.lastSeen).Seconds()
	bucket.tokens += elapsed * tbl.rate
	
	if bucket.tokens > float64(tbl.capacity) {
		bucket.tokens = float64(tbl.capacity)
	}
	
	bucket.lastSeen = now
	
	if bucket.tokens >= 1.0 {
		bucket.tokens--
		return true, 0
	}
	
	// Calculate time until next token (ensure minimum wait time)
	tokensNeeded := 1.0 - bucket.tokens
	waitTime := time.Duration(tokensNeeded/tbl.rate*1000) * time.Millisecond
	if waitTime <= 0 {
		waitTime = time.Millisecond
	}
	return false, waitTime
}

func (tbl *TokenBucketLimiter) Reset(key string) {
	tbl.mu.Lock()
	defer tbl.mu.Unlock()
	delete(tbl.buckets, key)
}

// SlidingWindowLimiter implements sliding window algorithm
type SlidingWindowLimiter struct {
	mu       sync.RWMutex
	windows  map[string]*slidingWindow
	requests int
	window   time.Duration
}

type slidingWindow struct {
	timestamps []time.Time
}

// NewSlidingWindowLimiter creates a new sliding window limiter
func NewSlidingWindowLimiter(requests int, window time.Duration) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		windows:  make(map[string]*slidingWindow),
		requests: requests,
		window:   window,
	}
}

func (swl *SlidingWindowLimiter) Allow(key string) (bool, time.Duration) {
	swl.mu.Lock()
	defer swl.mu.Unlock()
	
	now := time.Now()
	window, exists := swl.windows[key]
	
	if !exists {
		window = &slidingWindow{
			timestamps: make([]time.Time, 0),
		}
		swl.windows[key] = window
	}
	
	// Remove expired timestamps
	cutoff := now.Add(-swl.window)
	validTimestamps := window.timestamps[:0]
	for _, ts := range window.timestamps {
		if ts.After(cutoff) {
			validTimestamps = append(validTimestamps, ts)
		}
	}
	window.timestamps = validTimestamps
	
	if len(window.timestamps) < swl.requests {
		window.timestamps = append(window.timestamps, now)
		return true, 0
	}
	
	// Calculate time until oldest request expires
	oldestTime := window.timestamps[0]
	waitTime := swl.window - now.Sub(oldestTime)
	return false, waitTime
}

func (swl *SlidingWindowLimiter) Reset(key string) {
	swl.mu.Lock()
	defer swl.mu.Unlock()
	delete(swl.windows, key)
}

// FixedWindowLimiter implements fixed window algorithm
type FixedWindowLimiter struct {
	mu       sync.RWMutex
	windows  map[string]*fixedWindow
	requests int
	window   time.Duration
}

type fixedWindow struct {
	count     int
	windowStart time.Time
}

// NewFixedWindowLimiter creates a new fixed window limiter
func NewFixedWindowLimiter(requests int, window time.Duration) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		windows:  make(map[string]*fixedWindow),
		requests: requests,
		window:   window,
	}
}

func (fwl *FixedWindowLimiter) Allow(key string) (bool, time.Duration) {
	fwl.mu.Lock()
	defer fwl.mu.Unlock()
	
	now := time.Now()
	window, exists := fwl.windows[key]
	
	if !exists {
		window = &fixedWindow{
			count:       0,
			windowStart: now,
		}
		fwl.windows[key] = window
	}
	
	// Check if window has expired
	if now.Sub(window.windowStart) >= fwl.window {
		window.count = 0
		window.windowStart = now
	}
	
	if window.count < fwl.requests {
		window.count++
		return true, 0
	}
	
	// Calculate time until window resets
	waitTime := fwl.window - now.Sub(window.windowStart)
	return false, waitTime
}

func (fwl *FixedWindowLimiter) Reset(key string) {
	fwl.mu.Lock()
	defer fwl.mu.Unlock()
	delete(fwl.windows, key)
}

// extractKey extracts the rate limiting key from the request
func extractKey(r *http.Request, config Config) string {
	switch config.KeyExtractor {
	case ExtractorIP:
		return extractIPFromRequest(r)
	case ExtractorHeader:
		value := r.Header.Get(config.HeaderName)
		if value == "" {
			return extractIPFromRequest(r) // Fallback to IP
		}
		return value
	case ExtractorCustom:
		if config.CustomKeyFunc != nil {
			return config.CustomKeyFunc(r)
		}
		return extractIPFromRequest(r) // Fallback to IP
	default:
		return extractIPFromRequest(r)
	}
}

// extractIPFromRequest extracts the client IP from the request
func extractIPFromRequest(r *http.Request) string {
	// Check for forwarded headers first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain
		if ip := net.ParseIP(xff); ip != nil {
			return ip.String()
		}
	}
	
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		if ip := net.ParseIP(xri); ip != nil {
			return ip.String()
		}
	}
	
	// Fall back to RemoteAddr
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		if ip := net.ParseIP(host); ip != nil {
			return ip.String()
		}
	}
	
	return r.RemoteAddr
}

// createLimiter creates the appropriate limiter based on configuration
func createLimiter(config Config) Limiter {
	switch config.Algorithm {
	case TokenBucket:
		return NewTokenBucketLimiter(config.Requests, config.Window, config.Burst)
	case SlidingWindow:
		return NewSlidingWindowLimiter(config.Requests, config.Window)
	case FixedWindow:
		return NewFixedWindowLimiter(config.Requests, config.Window)
	default:
		return NewTokenBucketLimiter(config.Requests, config.Window, config.Burst)
	}
}

// Middleware creates rate limiting middleware
func Middleware(config Config) func(next http.Handler) http.Handler {
	if !config.Enabled {
		return func(next http.Handler) http.Handler {
			return next
		}
	}
	
	limiter := createLimiter(config)
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := extractKey(r, config)
			allowed, waitTime := limiter.Allow(key)
			
			if !allowed {
				// Set rate limit headers
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(config.Requests))
				w.Header().Set("X-RateLimit-Window", config.Window.String())
				w.Header().Set("X-RateLimit-Retry-After", strconv.Itoa(int(waitTime.Seconds())))
				
				// Return 429 Too Many Requests
				http.Error(w, fmt.Sprintf("Rate limit exceeded. Try again in %v", waitTime), http.StatusTooManyRequests)
				return
			}
			
			// Add rate limit info headers for successful requests
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(config.Requests))
			w.Header().Set("X-RateLimit-Window", config.Window.String())
			
			next.ServeHTTP(w, r)
		})
	}
}