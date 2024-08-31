package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/tomkaith13/dist-kv-store/internal/config"
	"github.com/tomkaith13/dist-kv-store/internal/router"
	"github.com/tomkaith13/dist-kv-store/internal/server"
)

func main() {

	zlogger := zerolog.New(os.Stderr).
		Level(zerolog.DebugLevel).
		With().
		Timestamp().
		Logger()

	config, err := config.LoadFromEnv()
	if err != nil {
		zlogger.Fatal().Err(err).Msg("failed to load env vars")
	}

	router := router.New()
	server := server.New(zlogger, router, config.Server)
	server.Run()

}
