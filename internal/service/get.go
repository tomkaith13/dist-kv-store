package service

import (
	"errors"
	"fmt"
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

	if !dkvService.ServiceConfig.Debug {
		dkvService.logger.Info().Msgf("using store raft.Stats: %+v", dkvService.raft.Stats())
	}

	key := chi.URLParam(r, "id")

	if len(key) > dkvService.ServiceConfig.KeyMaxLen {
		err := errors.New("key size exceeded")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Dummy set
	// dkvService.Set("asd", "def")

	// TODO: call store.Get()
	val, err := dkvService.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	resp := fmt.Sprintf("{ %q : %q }", key, val)
	w.Write([]byte(resp))
	w.WriteHeader(http.StatusOK)
}
