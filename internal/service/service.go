package service

import (
	"github.com/rs/zerolog"
)

type DKVService struct {
	logger zerolog.Logger
}

func New(logger zerolog.Logger) *DKVService {
	service := &DKVService{
		logger: logger,
	}
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

func (s *DKVService) Print() {
	s.logger.Info().Msg("Service")
}
