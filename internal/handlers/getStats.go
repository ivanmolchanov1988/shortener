// Package handlers содержит обработчики HTTP-запросов для работы с сервисом сокращения ссылок.
package handlers

import (
	"net"
	"net/http"
)

// statsResponse описывает JSON-ответ для эндпоинта /api/internal/stats.
type statsResponse struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}

// resolveClientIP получает IP клиента из X-Real-IP, X-Forwarded-For или RemoteAddr.
func resolveClientIP(r *http.Request) (net.IP, error) {
	// 1. Проверяем заголовок X-Real-IP
	ipStr := r.Header.Get("X-Real-IP")
	ip := net.ParseIP(ipStr)
	if ip != nil {
		return ip, nil
	}

	// 2. Проверяем заголовок X-Forwarded-For (цепочка IP через запятую)
	ips := r.Header.Get("X-Forwarded-For")
	if ips != "" {
		ipList := net.ParseIP(ips)
		if ipList != nil {
			return ipList, nil
		}
	}

	// 3. Если заголовки отсутствуют, берём IP из RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return nil, err
	}
	return net.ParseIP(host), nil
}
