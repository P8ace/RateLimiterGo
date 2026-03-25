package limiter

import (
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// func PerClientRateLimiter(next func(w http.ResponseWriter, r *http.Request)) http.Handler {
func PerClientRateLimiter(next http.HandlerFunc, max, burst int) http.Handler {

	type Client struct {
		limiter  *rate.Limiter
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
			clients[ip] = &Client{limiter: rate.NewLimiter(rate.Limit(max), burst)}
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
