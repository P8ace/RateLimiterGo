package limiter

import (
	"container/list"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"
)

// compile time interface implementation check
var _ RateLimiter = (*SlidingWindowLimiter)(nil)

type SlidingWindowLimiter struct {
	window  int64
	limit   int
	dequeue *list.List //dequeue - push_back, push_front in O(1) constant time
	mu      sync.Mutex
}

func NewSlidingWindowLimiter(window int64, limit int) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		window:  window,
		limit:   limit,
		dequeue: list.New(),
	}
}

func (rl *SlidingWindowLimiter) Allow() bool {
	now := time.Now()
	delta := now.Unix() - rl.window
	edgeTime := time.Unix(delta, 0)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// TODO: check if the for keyword has to be replaced with if
	// Remove outdated requests
	for rl.dequeue.Len() > 0 {
		front := rl.dequeue.Front()
		if front.Value.(time.Time).Before(edgeTime) { //if front value's time is before edgetime
			rl.dequeue.Remove(front)
		} else {
			break
		}
	}

	// check if the request can be added to the dequeue and allowed
	if rl.dequeue.Len() < rl.limit {
		rl.dequeue.PushBack(now)
		return true
	}

	return false
}

func SlidingWindowMiddleware(next http.HandlerFunc, ratelimiter RateLimiter) http.Handler {

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
		// get IP of the client
		ipAddr, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			slog.Error("Error retrieving the IP from Remote Addr", "Error:", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// rate limit the request
		mu.Lock()
		if _, ok := clients[ipAddr]; !ok {
			clients[ipAddr] = &Client{limiter: ratelimiter}
		}
		clients[ipAddr].lastSeen = time.Now()
		mu.Unlock()

		if !clients[ipAddr].limiter.Allow() {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		// call the handler function
		next(w, r)
	})
}
