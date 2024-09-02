package service

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/tomkaith13/dist-kv-store/internal/server"
)

type RegisterFollowerRequest struct {
	FollowerId   string `json:"follower_id"`
	FollowerAddr string `json:"follower_addr"`
}

var (
	RegistrationFailed error = errors.New("Registration of Follower failed")
)

func RegisterFollowerHandler(s *server.Server, w http.ResponseWriter, r *http.Request) {
	var reqBody RegisterFollowerRequest
	defer r.Body.Close()

	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		err := errors.New("unable to decode body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	store := s.GetStore()
	dkvService, ok := store.(*DKVService)
	if !ok {
		http.Error(w, "Unable to access store", http.StatusInternalServerError)
		return
	}

	err = dkvService.RegisterFollower(reqBody.FollowerId, reqBody.FollowerAddr)
	if err != nil {
		http.Error(w, RegistrationFailed.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("registered"))

}
