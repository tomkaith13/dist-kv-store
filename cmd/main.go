package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/tomkaith13/dist-kv-store/internal/config"
	"github.com/tomkaith13/dist-kv-store/internal/router"
	"github.com/tomkaith13/dist-kv-store/internal/server"
	"github.com/tomkaith13/dist-kv-store/internal/service"
)

func main() {

	// Initialize Logger
	zlogger := zerolog.New(os.Stderr).
		Level(zerolog.DebugLevel).
		With().
		Timestamp().
		Logger()

	// read configs using https://github.com/kelseyhightower/envconfig
	config, err := config.LoadFromEnv()
	if err != nil {
		zlogger.Fatal().Err(err).Msg("failed to load env vars")
	}

	kv_service := service.New(zlogger)
	r := router.New(config.Router, zlogger)
	// handler registration to the service
	// test route - for tests and checking if the setup works
	httpServer := server.New(zlogger, r.GetRouter(), config.Server, kv_service)
	httpServer.AddHandler(server.GET, "/hello", service.HelloHandler)
	// httpServer.AddHandler(server.GET, "/hello-long", service.HelloHandlerLong)

	// GET
	// server.AddHandler(router.GET, "/key", service.GetHandler)

	httpServer.Run()

}
