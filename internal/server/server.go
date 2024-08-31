package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

type Server struct {
	logger zerolog.Logger
	router *chi.Mux
	config Config
}

type Config struct {
	Address         string        `envconfig:"ADDRESS" default:"0.0.0.0:8000"`
	ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"5s"`
}

func New(logger zerolog.Logger, router *chi.Mux, config Config) *Server {
	return &Server{
		logger: logger,
		router: router,
		config: config,
	}
}

func (s *Server) Run() error {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	serverErrors := make(chan error, 1)

	api := &http.Server{
		Addr:    s.config.Address,
		Handler: s.router,
	}

	go func() {
		s.logger.Info().Msg("server listening on port 8888")
		serverErrors <- api.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server encountered an error: %w", err)
	case sig := <-shutdown:
		s.logger.Info().Msgf("server shutting down after receiving %+v", sig)
		ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
		defer cancel()

		if err := api.Shutdown(ctx); err != nil {
			_ = api.Close()
			return fmt.Errorf("server failed to shutdown gracefully: %w", err)
		}
	}
	return nil
}
