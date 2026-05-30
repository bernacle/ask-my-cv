package middleware

import (
	"testing"
	"time"
)

func TestRateLimiterAllowHonorsMinuteLimit(t *testing.T) {
	limiter := NewRateLimiter(2, 0)
	now := time.Date(2026, 5, 30, 12, 0, 0, 0, time.UTC)

	if !limiter.allow("127.0.0.1", now) {
		t.Fatal("first request should be allowed")
	}
	if !limiter.allow("127.0.0.1", now) {
		t.Fatal("second request should be allowed")
	}
	if limiter.allow("127.0.0.1", now) {
		t.Fatal("third request in same minute should be blocked")
	}
	if !limiter.allow("127.0.0.1", now.Add(time.Minute)) {
		t.Fatal("next minute should reset minute limit")
	}
}
