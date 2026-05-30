package middleware

import (
	"net"
	"net/http"
	"strings"
)

func ClientIP(r *http.Request) string {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		if first, _, ok := strings.Cut(forwardedFor, ","); ok {
			return strings.TrimSpace(first)
		}
		return strings.TrimSpace(forwardedFor)
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
