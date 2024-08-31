package service

import (
	"net/http"
	"time"
)

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("world!"))
	w.WriteHeader(http.StatusOK)
}

func HelloHandlerLong(w http.ResponseWriter, r *http.Request) {
	time.Sleep(15 * time.Second)
	w.Write([]byte("world!"))
	w.WriteHeader(http.StatusOK)
}
