package service

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/tomkaith13/dist-kv-store/internal/server"
)

func DelHandler(s *server.Server, w http.ResponseWriter, r *http.Request) {

	store := s.GetStore()
	dkvService, ok := store.(*DKVService)
	if !ok {
		http.Error(w, "Unable to access store", http.StatusInternalServerError)
		return
	}

	dkvService.logger.Info().Msgf("using service configs: %+v", dkvService.ServiceConfig)

	key := chi.URLParam(r, "id")

	_, err := dkvService.Delete(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	resp := "key deleted successfully"
	w.Write([]byte(resp))
	w.WriteHeader(http.StatusOK)
}
