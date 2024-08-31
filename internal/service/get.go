package service

import (
	"net/http"
)

func GetHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("world!"))
	w.WriteHeader(http.StatusOK)
}
