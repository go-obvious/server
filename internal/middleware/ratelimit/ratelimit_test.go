package ratelimit_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/go-obvious/server/internal/middleware/ratelimit"
)

func TestTokenBucketLimiter(t *testing.T) {
	tests := []struct {
		name        string
		requests    int
		window      time.Duration
		burst       int
		testReqs    int
		expectAllow int
	}{
		{
			name:        "Basic rate limiting",
			requests:    5,
			window:      time.Second,
			burst:       5,
			testReqs:    10,
			expectAllow: 5,
		},
		{
			name:        "Burst handling",
			requests:    2,
			window:      time.Second,
			burst:       5,
			testReqs:    7,
			expectAllow: 5,
		},
		{
			name:        "Zero burst uses requests as capacity",
			requests:    3,
			window:      time.Second,
			burst:       0,
			testReqs:    5,
			expectAllow: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := ratelimit.NewTokenBucketLimiter(tt.requests, tt.window, tt.burst)
			
			allowed := 0
			for i := 0; i < tt.testReqs; i++ {
				if ok, _ := limiter.Allow("test-key"); ok {
					allowed++
				}
			}
			
			assert.Equal(t, tt.expectAllow, allowed)
		})
	}
}

func TestTokenBucketLimiter_Refill(t *testing.T) {
	limiter := ratelimit.NewTokenBucketLimiter(2, time.Second, 2)
	
	// Exhaust tokens
	allowed, _ := limiter.Allow("test-key")
	assert.True(t, allowed)
	allowed, _ = limiter.Allow("test-key")
	assert.True(t, allowed)
	
	// Should be rate limited
	allowed, waitTime := limiter.Allow("test-key")
	assert.False(t, allowed)
	assert.Greater(t, waitTime, time.Duration(0))
	
	// Wait for refill
	time.Sleep(waitTime + 10*time.Millisecond)
	
	// Should allow at least one more request
	allowed, _ = limiter.Allow("test-key")
	assert.True(t, allowed)
}

func TestSlidingWindowLimiter(t *testing.T) {
	limiter := ratelimit.NewSlidingWindowLimiter(3, time.Second)
	
	// Allow first 3 requests
	for i := 0; i < 3; i++ {
		allowed, _ := limiter.Allow("test-key")
		assert.True(t, allowed, "Request %d should be allowed", i+1)
	}
	
	// 4th request should be denied
	allowed, waitTime := limiter.Allow("test-key")
	assert.False(t, allowed)
	assert.Greater(t, waitTime, time.Duration(0))
	
	// Wait for window to slide (wait for the full wait time)
	time.Sleep(waitTime + 10*time.Millisecond)
	
	// Should allow more requests
	allowed, _ = limiter.Allow("test-key")
	assert.True(t, allowed)
}

func TestFixedWindowLimiter(t *testing.T) {
	limiter := ratelimit.NewFixedWindowLimiter(2, time.Second)
	
	// Allow first 2 requests
	allowed, _ := limiter.Allow("test-key")
	assert.True(t, allowed)
	allowed, _ = limiter.Allow("test-key")
	assert.True(t, allowed)
	
	// 3rd request should be denied
	allowed, waitTime := limiter.Allow("test-key")
	assert.False(t, allowed)
	assert.Greater(t, waitTime, time.Duration(0))
	
	// Wait for window to reset
	time.Sleep(1100 * time.Millisecond)
	
	// Should allow requests in new window
	allowed, _ = limiter.Allow("test-key")
	assert.True(t, allowed)
}

func TestLimiterReset(t *testing.T) {
	tests := []struct {
		name    string
		limiter ratelimit.Limiter
	}{
		{
			name:    "TokenBucket",
			limiter: ratelimit.NewTokenBucketLimiter(1, time.Second, 1),
		},
		{
			name:    "SlidingWindow", 
			limiter: ratelimit.NewSlidingWindowLimiter(1, time.Second),
		},
		{
			name:    "FixedWindow",
			limiter: ratelimit.NewFixedWindowLimiter(1, time.Second),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Exhaust limit
			allowed, _ := tt.limiter.Allow("test-key")
			assert.True(t, allowed)
			
			// Should be rate limited
			allowed, _ = tt.limiter.Allow("test-key")
			assert.False(t, allowed)
			
			// Reset and try again
			tt.limiter.Reset("test-key")
			allowed, _ = tt.limiter.Allow("test-key")
			assert.True(t, allowed)
		})
	}
}

func TestRateLimitMiddleware_Disabled(t *testing.T) {
	config := ratelimit.Config{
		Enabled: false,
	}
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	
	middleware := ratelimit.Middleware(config)
	wrappedHandler := middleware(handler)
	
	// Make multiple requests - should all pass
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()
		
		wrappedHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	}
}

func TestRateLimitMiddleware_TokenBucket(t *testing.T) {
	config := ratelimit.Config{
		Enabled:      true,
		Requests:     3,
		Window:       time.Second,
		Burst:        3,
		Algorithm:    ratelimit.TokenBucket,
		KeyExtractor: ratelimit.ExtractorIP,
	}
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	
	middleware := ratelimit.Middleware(config)
	wrappedHandler := middleware(handler)
	
	// First 3 requests should pass
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		rr := httptest.NewRecorder()
		
		wrappedHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code, "Request %d should succeed", i+1)
		assert.Equal(t, "3", rr.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, "1s", rr.Header().Get("X-RateLimit-Window"))
	}
	
	// 4th request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil) 
	req.RemoteAddr = "127.0.0.1:12345"
	rr := httptest.NewRecorder()
	
	wrappedHandler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusTooManyRequests, rr.Code)
	assert.Contains(t, rr.Body.String(), "Rate limit exceeded")
	assert.NotEmpty(t, rr.Header().Get("X-RateLimit-Retry-After"))
}

func TestRateLimitMiddleware_HeaderExtractor(t *testing.T) {
	config := ratelimit.Config{
		Enabled:      true,
		Requests:     2,
		Window:       time.Second,
		Algorithm:    ratelimit.TokenBucket,
		KeyExtractor: ratelimit.ExtractorHeader,
		HeaderName:   "X-API-Key",
	}
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	
	middleware := ratelimit.Middleware(config)
	wrappedHandler := middleware(handler)
	
	// User 1 - should get 2 requests
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", "user1-key")
		rr := httptest.NewRecorder()
		
		wrappedHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	}
	
	// User 1 - 3rd request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "user1-key")
	rr := httptest.NewRecorder()
	
	wrappedHandler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusTooManyRequests, rr.Code)
	
	// User 2 - should still get requests (different key)
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "user2-key")
	rr = httptest.NewRecorder()
	
	wrappedHandler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRateLimitMiddleware_CustomExtractor(t *testing.T) {
	config := ratelimit.Config{
		Enabled:      true,
		Requests:     1,
		Window:       time.Second,
		Algorithm:    ratelimit.TokenBucket,
		KeyExtractor: ratelimit.ExtractorCustom,
		CustomKeyFunc: func(r *http.Request) string {
			return r.Header.Get("User-ID")
		},
	}
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	
	middleware := ratelimit.Middleware(config)
	wrappedHandler := middleware(handler)
	
	// User A - first request should pass
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-ID", "userA")
	rr := httptest.NewRecorder()
	
	wrappedHandler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	
	// User A - second request should be rate limited
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-ID", "userA")
	rr = httptest.NewRecorder()
	
	wrappedHandler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusTooManyRequests, rr.Code)
	
	// User B - should still get request (different user)
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-ID", "userB")
	rr = httptest.NewRecorder()
	
	wrappedHandler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRateLimitMiddleware_IPExtraction(t *testing.T) {
	tests := []struct {
		name           string
		remoteAddr     string
		forwardedFor   string
		realIP         string
		expectedKey    string
	}{
		{
			name:        "Basic RemoteAddr",
			remoteAddr:  "192.168.1.1:12345",
			expectedKey: "192.168.1.1",
		},
		{
			name:         "X-Forwarded-For header",
			remoteAddr:   "127.0.0.1:12345",
			forwardedFor: "203.0.113.1",
			expectedKey:  "203.0.113.1",
		},
		{
			name:        "X-Real-IP header",
			remoteAddr:  "127.0.0.1:12345",
			realIP:      "203.0.113.2",
			expectedKey: "203.0.113.2",
		},
		{
			name:         "X-Forwarded-For takes precedence",
			remoteAddr:   "127.0.0.1:12345",
			forwardedFor: "203.0.113.1",
			realIP:       "203.0.113.2",
			expectedKey:  "203.0.113.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ratelimit.Config{
				Enabled:      true,
				Requests:     1,
				Window:       time.Second,
				Algorithm:    ratelimit.TokenBucket,
				KeyExtractor: ratelimit.ExtractorIP,
			}
			
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			
			middleware := ratelimit.Middleware(config)
			wrappedHandler := middleware(handler)
			
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			
			if tt.forwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.forwardedFor)
			}
			if tt.realIP != "" {
				req.Header.Set("X-Real-IP", tt.realIP)
			}
			
			rr := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusOK, rr.Code)
			
			// Second request with same IP should be rate limited
			rr = httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusTooManyRequests, rr.Code)
		})
	}
}

func TestRateLimitMiddleware_Concurrency(t *testing.T) {
	config := ratelimit.Config{
		Enabled:      true,
		Requests:     100,
		Window:       time.Second,
		Algorithm:    ratelimit.TokenBucket,
		KeyExtractor: ratelimit.ExtractorIP,
	}
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	
	middleware := ratelimit.Middleware(config)
	wrappedHandler := middleware(handler)
	
	const numGoroutines = 50
	const requestsPerGoroutine = 5
	
	var wg sync.WaitGroup
	results := make(chan int, numGoroutines*requestsPerGoroutine)
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			
			for j := 0; j < requestsPerGoroutine; j++ {
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = fmt.Sprintf("192.168.1.%d:12345", goroutineID)
				rr := httptest.NewRecorder()
				
				wrappedHandler.ServeHTTP(rr, req)
				results <- rr.Code
			}
		}(i)
	}
	
	wg.Wait()
	close(results)
	
	okCount := 0
	rateLimitedCount := 0
	
	for code := range results {
		switch code {
		case http.StatusOK:
			okCount++
		case http.StatusTooManyRequests:
			rateLimitedCount++
		default:
			t.Errorf("Unexpected status code: %d", code)
		}
	}
	
	totalRequests := numGoroutines * requestsPerGoroutine
	assert.Equal(t, totalRequests, okCount+rateLimitedCount)
	
	// Since each goroutine uses a different IP, all requests should succeed
	assert.Equal(t, totalRequests, okCount)
	assert.Equal(t, 0, rateLimitedCount)
}

func TestRateLimitMiddleware_DifferentAlgorithms(t *testing.T) {
	algorithms := []ratelimit.Algorithm{
		ratelimit.TokenBucket,
		ratelimit.SlidingWindow,
		ratelimit.FixedWindow,
	}
	
	for _, algorithm := range algorithms {
		t.Run(string(algorithm), func(t *testing.T) {
			config := ratelimit.Config{
				Enabled:      true,
				Requests:     2,
				Window:       time.Second,
				Algorithm:    algorithm,
				KeyExtractor: ratelimit.ExtractorIP,
			}
			
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			
			middleware := ratelimit.Middleware(config)
			wrappedHandler := middleware(handler)
			
			// First 2 requests should pass
			for i := 0; i < 2; i++ {
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = "127.0.0.1:12345"
				rr := httptest.NewRecorder()
				
				wrappedHandler.ServeHTTP(rr, req)
				assert.Equal(t, http.StatusOK, rr.Code, "Request %d should succeed for %s", i+1, algorithm)
			}
			
			// 3rd request should be rate limited
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "127.0.0.1:12345"
			rr := httptest.NewRecorder()
			
			wrappedHandler.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusTooManyRequests, rr.Code, "Request should be rate limited for %s", algorithm)
		})
	}
}