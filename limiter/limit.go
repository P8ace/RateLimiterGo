package limiter

import (
	// "encoding/json
	"net/http"

	"golang.org/x/time/rate"
)

func RateLimiterFromXMiddleware(next func(w http.ResponseWriter, r *http.Request)) http.Handler {
	limiter := rate.NewLimiter(2, 4)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			// message := Message{
			// 	Status : "Request Failed",
			// 	Body: "API at capacity."
			// }

			w.WriteHeader(http.StatusTooManyRequests)
			// json.NewEncoder(w).Encode()
			return
		} else {
			next(w, r)
		}
	})
}
