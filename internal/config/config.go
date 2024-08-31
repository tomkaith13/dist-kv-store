package config

import (
	_ "github.com/joho/godotenv/autoload" // Autoload env vars from a .env file.
	"github.com/kelseyhightower/envconfig"
	"github.com/tomkaith13/dist-kv-store/internal/server"
)

// Config contains all the config
// parameters that this service uses.
type Config struct {
	Server server.Config `envconfig:"SERVER"`
}

// LoadFromEnv will load the env vars from the OS.
func LoadFromEnv() (*Config, error) {
	cfg := &Config{}
	err := envconfig.Process("", cfg)
	return cfg, err
}
