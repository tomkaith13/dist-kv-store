package service

import (
	"errors"
	"sync"

	"github.com/rs/zerolog"
)

type DKVService struct {
	logger        zerolog.Logger
	ServiceConfig Config
	mu            sync.Mutex
	kvmap         map[string]string
}

type Config struct {
	KeyMaxLen  int `envconfig:"KEY_MAX_LEN" default:"100"`
	ValMaxLen  int `envconfig:"VAL_MAX_LEN" default:"200"`
	MaxMapSize int `envconfig:"MAX_MAP_SIZE" default:"1000"`
}

func New(logger zerolog.Logger, config Config) *DKVService {
	service := &DKVService{
		logger:        logger,
		ServiceConfig: config,
	}

	service.kvmap = make(map[string]string)
	service.PrintConfigs()
	return service
}

func (s *DKVService) Get(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.kvmap[key]; !ok {
		return "", errors.New("key not found")
	}

	return s.kvmap[key], nil
}

func (s *DKVService) Set(key, val string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.kvmap[key]; ok {
		return "", errors.New("key already exists")
	}

	s.kvmap[key] = val
	return val, nil

}

func (s *DKVService) Delete(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.kvmap[key]; !ok {
		return "", errors.New("key not found")
	}

	delete(s.kvmap, key)
	return key, nil
}

func (s *DKVService) PrintConfigs() {
	s.logger.Info().Msg("--- KVService Config ---")
	s.logger.Info().Msgf("%+v", s.ServiceConfig)
	s.logger.Info().Msg("--- KVService Config ---")

}
