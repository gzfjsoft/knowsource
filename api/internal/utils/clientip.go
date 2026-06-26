package utils

import (
	"net"
	"net/http"
	"strings"
)

// ClientIP 优先 X-Forwarded-For / X-Real-Ip，否则 RemoteAddr
func ClientIP(r *http.Request) string {
	if r == nil {
		return ""
	}
	if x := r.Header.Get("X-Forwarded-For"); x != "" {
		return strings.TrimSpace(strings.Split(x, ",")[0])
	}
	if x := r.Header.Get("X-Real-Ip"); x != "" {
		return strings.TrimSpace(x)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
