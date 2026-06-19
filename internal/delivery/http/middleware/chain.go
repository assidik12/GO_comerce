package middleware

import (
	"net/http"
	"sync"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/time/rate"
)

// Chain wraps a handler with multiple middlewares.
func Chain(next httprouter.Handle, middlewares ...func(httprouter.Handle) httprouter.Handle) httprouter.Handle {
	for i := len(middlewares) - 1; i >= 0; i-- {
		next = middlewares[i](next)
	}
	return next
}

// RateLimit implements IP-based rate limiting.
func RateLimit(rps int, burst int) func(httprouter.Handle) httprouter.Handle {
	var limiters sync.Map

	return func(next httprouter.Handle) httprouter.Handle {
		return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			ip := r.RemoteAddr

			l, _ := limiters.LoadOrStore(ip, rate.NewLimiter(rate.Limit(rps), burst))
			limiter := l.(*rate.Limiter)

			if !limiter.Allow() {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}

			next(w, r, ps)
		}
	}
}
