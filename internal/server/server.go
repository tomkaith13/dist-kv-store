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

const (
	GET    = "GET"
	POST   = "POST"
	DELETE = "DEL"
)

type Server struct {
	logger zerolog.Logger
	router *chi.Mux
	config Config

	store DKVStore
}
type DKVStore interface {
	Get(key string) (string, error)
	Set(key string, val string) (string, error)
	Delete(key string) (string, error)
	RegisterFollower(followerId, followerAddr string) error
}
type Config struct {
	Address         string        `envconfig:"ADDRESS"`
	ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"5s"`
}

func New(logger zerolog.Logger, router *chi.Mux, config Config, store DKVStore) *Server {
	s := &Server{
		logger: logger,
		router: router,
		config: config,
		store:  store,
	}
	s.PrintConfigs()
	return s
}

func (s *Server) AddHandler(method string, route string, handlerFunc func(s *Server, w http.ResponseWriter, r *http.Request)) {
	s.logger.Info().Msgf("Registering Method: %s Route: %s", method, route)

	wrappedHandler := func(w http.ResponseWriter, r *http.Request) {
		handlerFunc(s, w, r)
	}
	switch method {
	case GET:
		s.router.Get(route, wrappedHandler)
	case POST:
		s.router.Post(route, wrappedHandler)
	case DELETE:
		s.router.Delete(route, wrappedHandler)
	default:
		s.logger.Fatal().Msg("Any other methods than GET, POST and DELETE are not allowed")
		return
	}
}

func (s *Server) PrintConfigs() {
	s.logger.Info().Msg("--- Server Config ---")
	s.logger.Info().Msgf("%+v", s.config)
	s.logger.Info().Msg("--- Server Config ---")
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
		s.logger.Info().Msg("server listening on " + s.config.Address)
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

func (s *Server) GetRouter() http.Handler {
	return s.router
}

func (s *Server) GetStore() DKVStore {
	return s.store
}
