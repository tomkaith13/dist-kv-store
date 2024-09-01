package service

import (
	"net/http"

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
	w.Write([]byte("called GET !"))
	w.WriteHeader(http.StatusOK)
}
