package service

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/tomkaith13/dist-kv-store/internal/server"
)

func GetHandler(s *server.Server, w http.ResponseWriter, r *http.Request) {

	store := s.GetStore()
	dkvService, ok := store.(*DKVService)
	if !ok {
		http.Error(w, "Unable to access store", http.StatusInternalServerError)
		return
	}

	dkvService.logger.Info().Msgf("using service configs: %+v", dkvService.ServiceConfig)

	key := chi.URLParam(r, "id")

	if len(key) > dkvService.ServiceConfig.KeyMaxLen {
		err := errors.New("key size exceeded")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: call store.Get()

	w.Write([]byte("called GET !"))
	w.WriteHeader(http.StatusOK)
}
