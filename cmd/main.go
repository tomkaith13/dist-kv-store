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

	// init service, router and finally init the server itself.
	kv_service := service.New(zlogger, config.Service)
	r := router.New(config.Router, zlogger)
	httpServer := server.New(zlogger, r.GetRouter(), config.Server, kv_service)

	// handler registration to the service
	httpServer.AddHandler(server.GET, "/hello", service.HelloHandler)
	httpServer.AddHandler(server.GET, "/key/{id}", service.GetHandler)
	httpServer.AddHandler(server.POST, "/key", service.SetHandler)

	// GET
	// server.AddHandler(router.GET, "/key", service.GetHandler)

	httpServer.Run()

}
