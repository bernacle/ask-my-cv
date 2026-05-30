package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/brunosimplicio/ask-my-cv/api/response"
	apperrors "github.com/brunosimplicio/ask-my-cv/pkg/errors"
)

type RateLimiter struct {
	perMinute int
	perDay    int
	mu        sync.Mutex
	visitors  map[string]*visitorState
}

type visitorState struct {
	minuteStart time.Time
	minuteCount int
	day         string
	dayCount    int
}

func NewRateLimiter(perMinute int, perDay int) *RateLimiter {
	return &RateLimiter{
		perMinute: perMinute,
		perDay:    perDay,
		visitors:  make(map[string]*visitorState),
	}
}

func (l *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		if !l.allow(ClientIP(r), time.Now().UTC()) {
			response.Error(w, r, apperrors.New(apperrors.CodeRateLimited, http.StatusTooManyRequests, "rate limit exceeded"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (l *RateLimiter) allow(key string, now time.Time) bool {
	if l.perMinute <= 0 && l.perDay <= 0 {
		return true
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	state := l.visitors[key]
	if state == nil {
		state = &visitorState{minuteStart: now.Truncate(time.Minute), day: now.Format("2006-01-02")}
		l.visitors[key] = state
	}

	currentMinute := now.Truncate(time.Minute)
	if !state.minuteStart.Equal(currentMinute) {
		state.minuteStart = currentMinute
		state.minuteCount = 0
	}
	currentDay := now.Format("2006-01-02")
	if state.day != currentDay {
		state.day = currentDay
		state.dayCount = 0
	}

	if l.perMinute > 0 && state.minuteCount >= l.perMinute {
		return false
	}
	if l.perDay > 0 && state.dayCount >= l.perDay {
		return false
	}
	state.minuteCount++
	state.dayCount++
	return true
}
