package service

import (
	"github.com/rs/zerolog"
)

type DKVService struct {
	logger zerolog.Logger
}

type DKVStore interface {
	Get(key string) (string, error)
	Set(key string, val string) error
	Delete(key string) error
}

func New(logger zerolog.Logger) *DKVService {
	service := &DKVService{
		logger: logger,
	}
	return service
}
