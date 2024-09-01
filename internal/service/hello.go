package service

import (
	"net/http"
	"time"

	"github.com/tomkaith13/dist-kv-store/internal/server"
)

func HelloHandler(s *server.Server, w http.ResponseWriter, r *http.Request) {
	if s == nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to init handlers"))
		return
	}
	w.Write([]byte("world!"))
	// s.PrintServer()
	w.WriteHeader(http.StatusOK)
}

func HelloHandlerLong(s *server.Server, w http.ResponseWriter, r *http.Request) {
	if s == nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to init handlers"))
		return
	}
	time.Sleep(10 * time.Second)
	w.Write([]byte("world!"))
	w.WriteHeader(http.StatusOK)
}
