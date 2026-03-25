package limiter

import (
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"
)

// func PerClientRateLimiter(next func(w http.ResponseWriter, r *http.Request)) http.Handler {
func PerClientRateLimiterMiddleware(next http.HandlerFunc, limiter RateLimiter) http.Handler {

	type Client struct {
		limiter  RateLimiter
		lastSeen time.Time
	}

	var mu sync.Mutex
	var clients = make(map[string]*Client)

	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > time.Minute*3 {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			slog.Error("Error retrieving the IP from Remote Addr", "Error:", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		mu.Lock()
		if _, ok := clients[ip]; !ok {
			clients[ip] = &Client{limiter: limiter}
		}
		clients[ip].lastSeen = time.Now()
		mu.Unlock()

		if !clients[ip].limiter.Allow() {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		} else {
			next(w, r)
		}
	})
}
