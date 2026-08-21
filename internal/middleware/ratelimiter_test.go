package middleware

import (
	"sync"
	"testing"
	"time"
)

func TestRateLimiterAllow(t *testing.T) {
	rl := NewRateLimiter(3, 1*time.Minute)

	if !rl.Allow("user1") {
		t.Error("first attempt should be allowed")
	}
	if !rl.Allow("user1") {
		t.Error("second attempt should be allowed")
	}
	if !rl.Allow("user1") {
		t.Error("third attempt should be allowed")
	}
	if rl.Allow("user1") {
		t.Error("fourth attempt should be blocked")
	}
}

func TestRateLimiterDifferentKeys(t *testing.T) {
	rl := NewRateLimiter(2, 1*time.Minute)

	rl.Allow("user1")
	rl.Allow("user1")
	if rl.Allow("user1") {
		t.Error("user1 should be blocked after 2 attempts")
	}

	if !rl.Allow("user2") {
		t.Error("user2 should not be affected by user1's attempts")
	}
}

func TestRateLimiterWindowReset(t *testing.T) {
	rl := NewRateLimiter(2, 50*time.Millisecond)

	rl.Allow("user1")
	rl.Allow("user1")
	if rl.Allow("user1") {
		t.Error("user1 should be blocked")
	}

	time.Sleep(60 * time.Millisecond)

	if !rl.Allow("user1") {
		t.Error("user1 should be allowed after window reset")
	}
}

func TestRateLimiterConcurrency(t *testing.T) {
	rl := NewRateLimiter(10, 1*time.Minute)
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rl.Allow("concurrent_user")
		}()
	}
	wg.Wait()
}

func TestRateLimiterZeroMax(t *testing.T) {
	rl := NewRateLimiter(0, 1*time.Minute)
	if rl.Allow("user1") {
		t.Error("with maxTry=0, no attempt should be allowed")
	}
}
