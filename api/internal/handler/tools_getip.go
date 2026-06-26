package handler

import (
	"net"
	"net/http"
	"strings"
)

func getClientIP(r *http.Request) string {
	// 尝试从 X-Real-IP 获取
	ip := r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	// 尝试从 X-Forwarded-For 获取
	ip = r.Header.Get("X-Forwarded-For")
	if ip != "" {
		ips := strings.Split(ip, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 使用 RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
