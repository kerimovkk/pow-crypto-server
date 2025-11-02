package server

import (
	"sync"
	"time"
)

// RateLimiter implements a sliding window rate limiter
type RateLimiter struct {
	requests      map[string][]time.Time // IP -> timestamps of requests
	mu            sync.RWMutex
	maxRequests   int           // Maximum requests allowed
	window        time.Duration // Time window
	cleanupTicker *time.Ticker
	cleanupStop   chan struct{}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxRequests int, window time.Duration, cleanupInterval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests:      make(map[string][]time.Time),
		maxRequests:   maxRequests,
		window:        window,
		cleanupTicker: time.NewTicker(cleanupInterval),
		cleanupStop:   make(chan struct{}),
	}

	// Start background cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// Allow checks if a request from the given IP should be allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	timestamps := rl.requests[ip]

	valid := make([]time.Time, 0)
	for _, t := range timestamps {
		if now.Sub(t) <= rl.window {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.maxRequests {
		return false
	}

	valid = append(valid, now)
	rl.requests[ip] = valid

	return true
}

// cleanupLoop periodically removes old entries from the map
func (rl *RateLimiter) cleanupLoop() {
	for {
		select {
		case <-rl.cleanupTicker.C:
			rl.cleanup()
		case <-rl.cleanupStop:
			return
		}
	}
}

// cleanup removes IPs with no recent requests
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, timestamps := range rl.requests {
		// Remove timestamps older than window
		valid := make([]time.Time, 0)
		for _, t := range timestamps {
			if now.Sub(t) <= rl.window {
				valid = append(valid, t)
			}
		}

		// If no valid timestamps, remove IP from map
		if len(valid) == 0 {
			delete(rl.requests, ip)
		} else {
			rl.requests[ip] = valid
		}
	}
}

// Stop stops the rate limiter cleanup goroutine
func (rl *RateLimiter) Stop() {
	rl.cleanupTicker.Stop()
	close(rl.cleanupStop)
}
