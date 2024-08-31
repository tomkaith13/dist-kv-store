package router

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/tomkaith13/dist-kv-store/internal/service"
)

const (
	// Global level timeout per request
	RequestTimeout time.Duration = 30 * time.Second
)

func New() *chi.Mux {
	r := chi.NewMux()

	r.Use(globalTimeoutMiddleware(RequestTimeout))
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// handler registration to the service
	r.Get("/hello", service.HelloHandler)

	return r
}

func globalTimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a new context with the specified timeout
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			// Wrap the request with the new context
			r = r.WithContext(ctx)

			// Channel to signal when the request is finished
			finished := make(chan struct{})

			go func() {
				next.ServeHTTP(w, r)
				close(finished)
			}()

			select {
			case <-finished:
				// Request finished normally
			case <-ctx.Done():
				// Timeout exceeded
				http.Error(w, ctx.Err().Error(), http.StatusGatewayTimeout)
			}
		})
	}
}
