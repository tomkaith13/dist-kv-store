package service

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/tomkaith13/dist-kv-store/internal/server"
)

type SetRequestBody struct {
	Key string `json:"key"`
	Val string `json:"value"`
}

func SetHandler(s *server.Server, w http.ResponseWriter, r *http.Request) {

	var reqBody SetRequestBody
	defer r.Body.Close()

	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		err := errors.New("Unable to decode body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	store := s.GetStore()
	dkvService, ok := store.(*DKVService)
	if !ok {
		http.Error(w, "Unable to access store", http.StatusInternalServerError)
		return
	}

	// Validations for key and val
	if len(reqBody.Key) > dkvService.ServiceConfig.KeyMaxLen {
		err := errors.New("key size exceeded")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(reqBody.Val) > dkvService.ServiceConfig.ValMaxLen {
		err := errors.New("value size exceeded")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(dkvService.kvmap) > dkvService.ServiceConfig.MaxMapSize {
		err := errors.New("max keys exceeded")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = dkvService.Set(reqBody.Key, reqBody.Val)
	if err != nil {
		if errors.Is(err, LeaderNotReady) {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	response := "key created successfully!"
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(response))
}
