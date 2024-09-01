package service

import (
	"sync"

	"github.com/rs/zerolog"
)

type DKVService struct {
	logger zerolog.Logger
	config Config
	mu     sync.Mutex
	kvmap  map[string]string
}

type Config struct {
	KeyMaxLen int `envconfig:"KEY_MAX_LEN" default:"100"`
	ValMaxLen int `envconfig:"VAL_MAX_LEN" default:"200"`
}

func New(logger zerolog.Logger, config Config) *DKVService {
	service := &DKVService{
		logger: logger,
		config: config,
	}
	service.PrintConfigs()
	return service
}

func (s *DKVService) Get(key string) (string, error) {
	return "", nil
}

func (s *DKVService) Set(key, val string) string {
	return ""
}

func (s *DKVService) Delete(key string) string {
	return ""
}

func (s *DKVService) PrintConfigs() {
	s.logger.Info().Msg("--- KVService Config ---")
	s.logger.Info().Msgf("%+v", s.config)
	s.logger.Info().Msg("--- KVService Config ---")

}
