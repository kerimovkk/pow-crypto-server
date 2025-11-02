package server

import (
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	// Config: 3 requests per 1 second
	rl := NewRateLimiter(3, time.Second, time.Minute)
	defer rl.Stop()

	ip := "192.168.1.1"

	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		if !rl.Allow(ip) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 4th request should be blocked
	if rl.Allow(ip) {
		t.Error("4th request should be blocked")
	}

	// Wait for window to expire
	time.Sleep(1100 * time.Millisecond)

	// Should allow again after window expires
	if !rl.Allow(ip) {
		t.Error("Request after window should be allowed")
	}
}

func TestRateLimiter_MultipleIPs(t *testing.T) {
	rl := NewRateLimiter(2, time.Second, time.Minute)
	defer rl.Stop()

	ip1 := "192.168.1.1"
	ip2 := "192.168.1.2"

	// IP1: 2 requests allowed
	if !rl.Allow(ip1) {
		t.Error("IP1 request 1 should be allowed")
	}
	if !rl.Allow(ip1) {
		t.Error("IP1 request 2 should be allowed")
	}

	// IP1: 3rd blocked
	if rl.Allow(ip1) {
		t.Error("IP1 request 3 should be blocked")
	}

	// IP2: independent limit, should be allowed
	if !rl.Allow(ip2) {
		t.Error("IP2 request 1 should be allowed")
	}
	if !rl.Allow(ip2) {
		t.Error("IP2 request 2 should be allowed")
	}

	// IP2: 3rd blocked
	if rl.Allow(ip2) {
		t.Error("IP2 request 3 should be blocked")
	}
}

func TestRateLimiter_SlidingWindow(t *testing.T) {
	// Config: 3 requests per 2 seconds
	rl := NewRateLimiter(3, 2*time.Second, time.Minute)
	defer rl.Stop()

	ip := "192.168.1.1"

	// t=0s: 3 requests
	for i := 0; i < 3; i++ {
		if !rl.Allow(ip) {
			t.Errorf("Request %d at t=0 should be allowed", i+1)
		}
	}

	// t=0s: 4th blocked
	if rl.Allow(ip) {
		t.Error("4th request at t=0 should be blocked")
	}

	// t=1.1s: wait 1.1s, oldest request still in window
	time.Sleep(1100 * time.Millisecond)

	// Still blocked (3 requests in last 2 seconds)
	if rl.Allow(ip) {
		t.Error("Request at t=1.1s should be blocked")
	}

	// t=2.1s: wait another 1s, first request expired
	time.Sleep(1000 * time.Millisecond)

	// Should be allowed now (only 2 requests in last 2 seconds)
	if !rl.Allow(ip) {
		t.Error("Request at t=2.1s should be allowed")
	}
}
