package router

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
)

type Config struct {
	RequestTimeout time.Duration `envconfig:"REQUEST_TIMEOUT" default:"30s"`
}

type Router struct {
	config    Config
	chiRouter *chi.Mux
	logger    zerolog.Logger
}

func New(config Config, logger zerolog.Logger) *Router {
	r := &Router{
		config:    config,
		chiRouter: chi.NewRouter(),
		logger:    logger,
	}
	r.PrintConfigs()
	r.setup()
	return r
}

func (r *Router) setup() {
	r.chiRouter.Use(globalTimeoutMiddleware(r.config.RequestTimeout, r.logger))
	r.chiRouter.Use(middleware.Logger)
	r.chiRouter.Use(middleware.Recoverer)
}

func (r *Router) GetRouter() *chi.Mux {
	return r.chiRouter
}

func (r *Router) PrintConfigs() {
	r.logger.Info().Msg("--- Router Config ---")
	r.logger.Info().Msgf("%+v", r.config)
	r.logger.Info().Msg("--- Router Config ---")
}

func globalTimeoutMiddleware(timeout time.Duration, logger zerolog.Logger) func(http.Handler) http.Handler {
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
				logger.Info().Msg("Request timed out! Check .env file for the value")
				http.Error(w, ctx.Err().Error(), http.StatusGatewayTimeout)
				return
			}
		})
	}
}
